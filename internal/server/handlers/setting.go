package handlers

import (
	"context"
	"net/http"

	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func init() {

	router.NewGroupRouter("/api/v1/setting").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getSetting),
		).
		AddRoute(
			router.NewRoute("", router.PUT).
				Handle(updateSetting),
		)
}

// getSetting 获取配置项
// @Summary 获取配置项
// @Description 获取系统所有配置项，支持按分组过滤和关键字搜索
// @Tags 配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group query string false "分组名称"
// @Success 200 {object} resp.ResponseStruct{data=[]setting.Setting} "获取成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/setting [get]
func getSetting(c *gin.Context) {
	result, err := op.GetAllSetting(context.Background())
	if err != nil {
		log.Errorf("Failed to get all setting: %v", err)
		resp.Error(c, http.StatusInternalServerError, "failed to get all setting")
		return
	}
	resp.Success(c, result)
}

// updateSetting 更新配置项
// @Summary 更新配置项
// @Description 根据请求数据中的ID批量更新配置项的值和描述
// @Tags 配置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []setting.Setting  true "更新配置项请求"
// @Success 200 {object} resp.ResponseStruct "更新成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/setting [put]
func updateSetting(c *gin.Context) {
	var req []setting.Setting
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}

	err := op.UpdateSetting(context.Background(), &req)
	if err != nil {
		log.Errorf("Failed to update config: %v", err)
		resp.Error(c, http.StatusInternalServerError, "failed to update config")
		return
	}

	resp.Success(c, nil)
}
