package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bestruirui/bestsub/internal/core/cron"
	"github.com/bestruirui/bestsub/internal/core/node"
	"github.com/bestruirui/bestsub/internal/database/op"
	"github.com/bestruirui/bestsub/internal/models/sub"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

// 防重复提交锁（基于订阅名称）
var (
	creatingSubs   = make(map[string]time.Time)
	creatingSubsMu sync.Mutex
)

// isDuplicateRequest 检查是否重复提交（5秒内相同名称视为重复）
func isDuplicateRequest(name string) bool {
	creatingSubsMu.Lock()
	defer creatingSubsMu.Unlock()

	if lastTime, exists := creatingSubs[name]; exists {
		if time.Since(lastTime) < 5*time.Second {
			return true
		}
	}
	creatingSubs[name] = time.Now()
	return false
}

// cleanupCreatingSubs 清理过期的创建记录
func cleanupCreatingSubs() {
	creatingSubsMu.Lock()
	defer creatingSubsMu.Unlock()

	for name, lastTime := range creatingSubs {
		if time.Since(lastTime) > 10*time.Second {
			delete(creatingSubs, name)
		}
	}
}

func init() {
	router.NewGroupRouter("/api/v1/sub").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.POST).
				Handle(createSub),
		).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getSubs),
		).
		AddRoute(
			router.NewRoute("/:id", router.PUT).
				Handle(updateSub),
		).
		AddRoute(
			router.NewRoute("/:id", router.DELETE).
				Handle(deleteSub),
		).
		AddRoute(
			router.NewRoute("/refresh/:id", router.POST).
				Handle(refreshSub),
		).
		AddRoute(
			router.NewRoute("/log", router.GET).
				Handle(getSubRunLog),
		).
		AddRoute(
			router.NewRoute("/batch", router.POST).
				Handle(batchCreateSub),
		)
}

// createSub 创建订阅链接
// @Summary 创建订阅链接
// @Description 创建单个订阅链接，订阅获取将异步执行，立即返回
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body sub.Request true "创建订阅链接请求"
// @Success 200 {object} resp.ResponseStruct{data=sub.Response} "创建成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub [post]
func createSub(c *gin.Context) {
	cleanupCreatingSubs()

	var req sub.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}

	// 防重复提交检查
	if isDuplicateRequest(req.Name) {
		log.Warnf("duplicate sub creation request: %s", req.Name)
		resp.Error(c, http.StatusTooManyRequests, "订阅正在创建中，请勿重复提交")
		return
	}

	subData := req.GenData(0)
	if err := op.CreateSub(c.Request.Context(), &subData); err != nil {
		log.Errorf("failed to create sub: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	cron.FetchAdd(&subData)

	// 异步执行订阅获取，避免阻塞响应
	if subData.Enable {
		go func(subID uint16) {
			cron.FetchRun(subID)
		}(subData.ID)
	}

	// 立即返回，不等待获取完成
	respData := subData.GenResponse(cron.FetchStatus(subData.ID, subData.Enable), node.GetSubInfo(subData.ID))
	resp.Success(c, respData)
}

// getSubs 获取订阅链接
// @Summary 获取订阅链接
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query int true "链接ID"
// @Success 200 {object} resp.ResponseStruct{data=[]sub.Response} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub [get]
func getSubs(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		subList, err := op.GetSubList(c.Request.Context())
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		var respSubList = make([]sub.Response, len(subList))
		for i := range subList {
			respSubList[i] = subList[i].GenResponse(cron.FetchStatus(subList[i].ID, subList[i].Enable), node.GetSubInfo(subList[i].ID))
		}
		resp.Success(c, respSubList)
	} else {
		id, err := strconv.ParseUint(idStr, 10, 16)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		subData, err := op.GetSubByID(c.Request.Context(), uint16(id))
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		var respSub = [1]sub.Response{}
		respSub[0] = subData.GenResponse(cron.FetchStatus(subData.ID, subData.Enable), node.GetSubInfo(subData.ID))
		resp.Success(c, respSub)
	}
}

// updateSub 更新订阅链接
// @Summary 更新订阅链接
// @Description 根据请求体中的ID更新订阅链接信息
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订阅链接ID"
// @Param request body sub.Request true "更新订阅链接请求"
// @Success 200 {object} resp.ResponseStruct{data=sub.Response} "更新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "订阅链接不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub/{id} [put]
func updateSub(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	var req sub.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	subData := req.GenData(uint16(id))
	if err := op.UpdateSub(c.Request.Context(), &subData); err != nil {
		log.Errorf("failed to update sub: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := cron.FetchUpdate(&subData); err != nil {
		log.Errorf("failed to update sub: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	respData := subData.GenResponse(cron.FetchStatus(subData.ID, subData.Enable), node.GetSubInfo(subData.ID))
	resp.Success(c, respData)
}

// deleteSub 删除订阅链接
// @Summary 删除订阅链接
// @Description 根据ID删除单个订阅链接
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订阅链接ID"
// @Success 200 {object} resp.ResponseStruct "删除成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "订阅链接不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub/{id} [delete]
func deleteSub(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	if err := op.DeleteSub(c.Request.Context(), uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := cron.FetchRemove(uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	node.DeleteBySubId(uint16(id))
	resp.Success(c, nil)
}

// refreshSub 手动刷新订阅
// @Summary 手动刷新订阅
// @Description 根据ID手动刷新单个订阅
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "订阅链接ID"
// @Success 200 {object} resp.ResponseStruct{data=sub.Result} "刷新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "订阅链接不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub/refresh/{id} [post]
func refreshSub(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	go cron.FetchRun(uint16(id))
	resp.Success(c, nil)
}

// batchCreateSub 批量创建订阅链接
// @Summary 批量创建订阅链接
// @Description 批量创建多个订阅链接
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []sub.Request true "批量创建订阅链接请求"
// @Success 200 {object} resp.ResponseStruct{data=[]sub.Response} "创建成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub/batch [post]
func batchCreateSub(c *gin.Context) {
	var reqs []sub.Request
	if err := c.ShouldBindJSON(&reqs); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	if len(reqs) == 0 {
		resp.ErrorBadRequest(c)
		return
	}

	subs := make([]*sub.Data, len(reqs))
	for i, req := range reqs {
		subData := req.GenData(0)
		subs[i] = &subData
	}

	if err := op.BatchCreateSub(c.Request.Context(), subs); err != nil {
		log.Errorf("failed to batch create subs: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	for _, subData := range subs {
		cron.FetchAdd(subData)
	}

	respData := make([]sub.Response, len(subs))
	for i, subData := range subs {
		respData[i] = subData.GenResponse(cron.FetchStatus(subData.ID, subData.Enable), node.GetSubInfo(subData.ID))
	}
	resp.Success(c, respData)
}
