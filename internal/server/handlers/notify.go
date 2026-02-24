package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/evilCYH/NodeHub/internal/database/op"
	notifyModel "github.com/evilCYH/NodeHub/internal/models/notify"
	"github.com/evilCYH/NodeHub/internal/modules/notify"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/notify").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/channel", router.GET).
				Handle(getNotifyChannel),
		).
		AddRoute(
			router.NewRoute("/channel/config", router.GET).
				Handle(getNotifyChannelConfig),
		).
		AddRoute(
			router.NewRoute("/name", router.GET).
				Handle(getNotifyNameAndID),
		).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getNotifyList),
		).
		AddRoute(
			router.NewRoute("", router.POST).
				Handle(createNotify),
		).
		AddRoute(
			router.NewRoute("", router.PUT).
				Handle(updateNotify),
		).
		AddRoute(
			router.NewRoute("", router.DELETE).
				Handle(deleteNotify),
		).
		AddRoute(
			router.NewRoute("/test", router.POST).
				Handle(testNotify),
		).
		AddRoute(
			router.NewRoute("/template", router.GET).
				Handle(getTemplates),
		).
		AddRoute(
			router.NewRoute("/template", router.PUT).
				Handle(updateTemplate),
		)
}

// getNotifyConfig 获取通知渠道
// @Summary 获取通知渠道
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]string} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/channel [get]
func getNotifyChannel(c *gin.Context) {
	resp.Success(c, notify.GetChannels())
}

// getNotifyChannelConfig 获取渠道配置
// @Summary 获取渠道配置
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel query string false "渠道"
// @Success 200 {object} resp.ResponseStruct{data=map[string][]notify.Desc} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/channel/config [get]
func getNotifyChannelConfig(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		resp.Success(c, notify.GetInfoMap())
	} else {
		resp.Success(c, notify.GetInfoMap()[channel])
	}
}

// getNotifyList 获取通知
// @Summary 获取通知
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]notifyModel.Response} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify [get]
func getNotifyList(c *gin.Context) {
	notifyList, err := op.GetNotifyList()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, notifyList)
}

// getNotifyNameAndID 获取通知名称和ID
// @Summary 获取通知名称和ID
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]notifyModel.NameAndID} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/name [get]
func getNotifyNameAndID(c *gin.Context) {
	notifyList, err := op.GetNotifyList()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	nameAndIDList := make([]notifyModel.NameAndID, len(notifyList))
	for i, notify := range notifyList {
		nameAndIDList[i] = notifyModel.NameAndID{
			ID:   notify.ID,
			Name: notify.Name,
		}
	}
	resp.Success(c, nameAndIDList)
}

// createNotify 创建通知
// @Summary 创建通知
// @Description 创建单个通知
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body notifyModel.Request true "创建通知请求"
// @Success 200 {object} resp.ResponseStruct{data=notifyModel.Response} "创建成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify [post]
func createNotify(c *gin.Context) {
	var req notifyModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	notifyData := req.GenData(0)
	types := notify.GetChannels()
	if !slices.Contains(types, req.Type) {
		resp.Error(c, http.StatusBadRequest, fmt.Sprintf("通知类型 %s 不存在", req.Type))
		return
	}
	if err := op.CreateNotify(c.Request.Context(), &notifyData); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	log.Infof("Notify config %d created by from %s", notifyData.ID, c.ClientIP())
	resp.Success(c, notifyData.GenResponse())
}

// testNotify 测试通知
// @Summary 测试通知
// @Description 测试单个通知
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body notifyModel.Request true "测试通知请求"
// @Success 200 {object} resp.ResponseStruct "测试成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/test [post]
func testNotify(c *gin.Context) {
	var req notifyModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	types := notify.GetChannels()
	if !slices.Contains(types, req.Type) {
		resp.Error(c, http.StatusBadRequest, fmt.Sprintf("通知类型 %s 不存在", req.Type))
		return
	}
	notifyData := req.GenData(0)
	notify, err := notify.Get(notifyData.Type, notifyData.Config)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	err = notify.Init()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	var buf bytes.Buffer
	buf.WriteString("test")
	err = notify.Send("test", &buf)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	log.Infof("Notify config %s tested by from %s", notifyData.Type, c.ClientIP())
	resp.Success(c, nil)
}

// updateNotify 更新通知
// @Summary 更新通知
// @Description 根据请求体中的ID更新通知信息
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query int true "通知ID"
// @Param request body notifyModel.Request true "更新通知请求"
// @Success 200 {object} resp.ResponseStruct{data=notifyModel.Response} "更新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "通知配置不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify [put]
func updateNotify(c *gin.Context) {
	var req notifyModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	idParam := c.Query("id")
	if idParam == "" {
		resp.ErrorBadRequest(c)
		return
	}
	id, err := strconv.ParseUint(idParam, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	notifyData := req.GenData(uint16(id))
	if err := op.UpdateNotify(c.Request.Context(), &notifyData); err != nil {
		log.Errorf("Update notify config %d failed: %v", id, err)
		resp.Error(c, http.StatusInternalServerError, "update notify config failed")
		return
	}
	log.Infof("Notify config %d updated by from %s", id, c.ClientIP())
	resp.Success(c, notifyData.GenResponse())
}

// deleteNotify 删除通知
// @Summary 删除通知
// @Description 根据ID删除单个通知
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query int true "通知ID"
// @Success 200 {object} resp.ResponseStruct "删除成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "通知不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify [delete]
func deleteNotify(c *gin.Context) {
	idParam := c.Query("id")
	if idParam == "" {
		resp.ErrorBadRequest(c)
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	if err := op.DeleteNotify(c.Request.Context(), uint16(id)); err != nil {
		log.Errorf("Delete notify config %d failed: %v", id, err)
		resp.Error(c, http.StatusInternalServerError, "delete notify config failed")
		return
	}
	log.Infof("Notify config %d deleted by from %s", id, c.ClientIP())

	resp.Success(c, nil)
}

// getTemplates 获取通知模板
// @Summary 获取通知模板
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]notifyModel.Template} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/template [get]
func getTemplates(c *gin.Context) {
	notifyTemplateList, err := op.GetNotifyTemplateList()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, notifyTemplateList)
}

// updateTemplate 更新通知模板
// @Summary 更新通知模板
// @Description 根据请求体中的ID更新通知模板信息
// @Tags 通知
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body notifyModel.Template true "更新通知模板请求"
// @Success 200 {object} resp.ResponseStruct{data=notifyModel.Template} "更新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "通知模板不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/notify/template [put]
func updateTemplate(c *gin.Context) {
	var req notifyModel.Template
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := op.UpdateNotifyTemplate(c.Request.Context(), &req); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	log.Infof("Notify template %s updated by from %s", req.Type, c.ClientIP())
	resp.Success(c, req)
}
