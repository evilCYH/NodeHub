package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/evilCYH/NodeHub/internal/core/check"
	"github.com/evilCYH/NodeHub/internal/core/cron"
	"github.com/evilCYH/NodeHub/internal/database/op"
	checkModel "github.com/evilCYH/NodeHub/internal/models/check"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func init() {

	router.NewGroupRouter("/api/v1/check").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/type", router.GET).
				Handle(getCheckTypes),
		).
		AddRoute(
			router.NewRoute("", router.POST).
				Handle(createCheck),
		).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getCheck),
		).
		AddRoute(
			router.NewRoute("/:id", router.PUT).
				Handle(updateCheck),
		).
		AddRoute(
			router.NewRoute("/:id", router.DELETE).
				Handle(deleteCheck),
		).
		AddRoute(
			router.NewRoute("/:id/run", router.POST).
				Handle(runCheck),
		).
		AddRoute(
			router.NewRoute("/:id/stop", router.POST).
				Handle(stopCheck),
		)
}

// getCheckTypes 获取检测类型
// @Summary 获取检测类型
// @Description 获取检测类型
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=map[string][]check.Desc} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check/type [get]
func getCheckTypes(c *gin.Context) {
	resp.Success(c, check.GetInfoMap())
}

// createCheck 创建检测
// @Summary 创建检测
// @Description 创建单个检测
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body checkModel.Request true "创建检测请求"
// @Success 200 {object} resp.ResponseStruct{data=checkModel.Response} "创建成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check [post]
func createCheck(c *gin.Context) {
	var req checkModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	checkData := req.GenData()
	if err := op.CreateCheck(c.Request.Context(), &checkData); err != nil {
		log.Errorf("failed to create check: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	cron.CheckAdd(&checkData)
	resp.Success(c, checkData.GenResponse(cron.CheckStatus(checkData.ID)))
}

// getCheck 获取检测列表
// @Summary 获取检测列表
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id query int true "检测ID"
// @Success 200 {object} resp.ResponseStruct{data=[]checkModel.Response} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check [get]
func getCheck(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		checkList, err := op.GetCheckList()
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		var respCheckList = make([]checkModel.Response, len(checkList))
		for i := range checkList {
			respCheckList[i] = checkList[i].GenResponse(cron.CheckStatus(checkList[i].ID))
		}
		resp.Success(c, respCheckList)
	} else {
		id, err := strconv.ParseUint(idStr, 10, 16)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		check, err := op.GetCheckByID(uint16(id))
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		var respCheck = make([]checkModel.Response, 1)
		respCheck[0] = check.GenResponse(cron.CheckStatus(check.ID))
		resp.Success(c, respCheck)
	}
}

// updateCheck 更新检测
// @Summary 更新检测
// @Description 根据请求体中的ID更新检测信息
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "检测ID"
// @Param request body checkModel.Request true "更新检测请求"
// @Success 200 {object} resp.ResponseStruct{data=checkModel.Response} "更新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "检测不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check/{id} [put]
func updateCheck(c *gin.Context) {
	var req checkModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	idStr := c.Param("id")
	if idStr == "" {
		resp.ErrorBadRequest(c)
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	checkData := req.GenData()
	checkData.ID = uint16(id)
	if err := op.UpdateCheck(c.Request.Context(), &checkData); err != nil {
		log.Errorf("failed to update check: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := cron.CheckUpdate(&checkData); err != nil {
		log.Errorf("failed to update check: %v", err)
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, checkData.GenResponse(cron.CheckStatus(checkData.ID)))
}

// deleteCheck 删除检测
// @Summary 删除检测
// @Description 根据ID删除单个检测
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "检测ID"
// @Success 200 {object} resp.ResponseStruct "删除成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "检测不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check/{id} [delete]
func deleteCheck(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		resp.Error(c, http.StatusBadRequest, "check id is required")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}

	if err := op.DeleteCheck(c.Request.Context(), uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, "failed to delete check")
		return
	}
	if err := cron.CheckRemove(uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if err := log.DeleteLog(fmt.Sprintf("check/%d", id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, nil)
}

// runCheck 手动运行检测
// @Summary 手动运行检测
// @Description 手动触发检测执行
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "检测ID"
// @Success 200 {object} resp.ResponseStruct "运行成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "检测不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check/{id}/run [post]
func runCheck(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		resp.Error(c, http.StatusBadRequest, "check id is required")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}

	if err := cron.CheckRun(uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Debugf("Check %d manually run from %s", id, c.ClientIP())

	resp.Success(c, nil)
}

// stopCheck 停止检测
// @Summary 停止检测
// @Description 停止正在运行的检测
// @Tags 检测
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "检测ID"
// @Success 200 {object} resp.ResponseStruct "停止成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "检测不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/check/{id}/stop [post]
func stopCheck(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		resp.Error(c, http.StatusBadRequest, "task id is required")
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}

	if err := cron.CheckStop(uint16(id)); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	log.Infof("Check %d stopped from %s", id, c.ClientIP())

	resp.Success(c, nil)
}
