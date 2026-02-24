package handlers

import (
	"net/http"

	"github.com/evilCYH/NodeHub/internal/core/update"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/update").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(latest),
		).
		AddRoute(
			router.NewRoute("/:name", router.POST).
				Handle(updateFunc),
		)
}

// latest 最新版本
// @Summary 最新版本
// @Description 获取程序最新版本信息
// @Tags 更新
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=map[string]update.LatestInfo} "获取成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/update [get]
func latest(c *gin.Context) {
	latestInfo := make(map[string]update.LatestInfo, 1)
	nodehub, err := update.GetLatestNodeHubInfo()
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	latestInfo["nodehub"] = *nodehub
	resp.Success(c, latestInfo)
}

// update 更新
// @Summary 更新
// @Description 更新程序
// @Tags 更新
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=string} "获取成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/update/:name [post]
func updateFunc(c *gin.Context) {
	name := c.Param("name")
	switch name {
	case "nodehub":
		err := update.UpdateCore()
		if err != nil {
			resp.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	default:
		resp.ErrorBadRequest(c)
	}
	resp.Success(c, nil)
}
