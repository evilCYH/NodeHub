package handlers

import (
	"net/http"
	"strconv"

	"github.com/evilCYH/NodeHub/internal/database/op"
	storageModel "github.com/evilCYH/NodeHub/internal/models/storage"
	"github.com/evilCYH/NodeHub/internal/modules/storage"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/storage").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.POST).
				Handle(createStorage),
		).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getStorage),
		).
		AddRoute(
			router.NewRoute("/:id", router.PUT).
				Handle(updateStorage),
		).
		AddRoute(
			router.NewRoute("/:id", router.DELETE).
				Handle(deleteStorage),
		).
		AddRoute(
			router.NewRoute("/channel", router.GET).
				Handle(getStorageChannel),
		).
		AddRoute(
			router.NewRoute("/channel/config", router.GET).
				Handle(getStorageChannelConfig),
		)
}

// @Summary 创建存储
// @Description 创建存储
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body storageModel.Request true "存储配置数据"
// @Success 200 {object} resp.ResponseStruct{data=storageModel.Response} "创建成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage [post]
func createStorage(c *gin.Context) {
	var req storageModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("createStorage: %v", err)
		resp.ErrorBadRequest(c)
		return
	}
	data := req.GenData(0)
	if err := op.CreateStorage(c.Request.Context(), &data); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, data.GenResponse())
}

// @Summary 获取存储
// @Description 获取存储
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]storageModel.Response} "获取成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage [get]
func getStorage(c *gin.Context) {
	storages, err := op.GetStorageList(c.Request.Context())
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	result := make([]storageModel.Response, len(storages))
	for i, v := range storages {
		result[i] = v.GenResponse()
	}
	resp.Success(c, result)
}

// getStorageChannel 获取存储渠道
// @Summary 获取存储渠道
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} resp.ResponseStruct{data=[]string} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage/channel [get]
func getStorageChannel(c *gin.Context) {
	channels := make([]string, 0, len(storage.GetInfoMap()))
	for channel := range storage.GetInfoMap() {
		channels = append(channels, channel)
	}
	resp.Success(c, channels)
}

// getStorageChannelConfig 获取渠道配置
// @Summary 获取渠道配置
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param channel query string false "渠道"
// @Success 200 {object} resp.ResponseStruct{data=map[string][]storage.Desc} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage/channel/config [get]
func getStorageChannelConfig(c *gin.Context) {
	channel := c.Query("channel")
	if channel == "" {
		resp.Success(c, storage.GetInfoMap())
	} else {
		resp.Success(c, storage.GetInfoMap()[channel])
	}
}

// @Summary 更新存储
// @Description 更新存储
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "存储ID"
// @Param data body storageModel.Request true "存储配置数据"
// @Success 200 {object} resp.ResponseStruct{data=storageModel.Response} "更新成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage/{id} [put]
func updateStorage(c *gin.Context) {
	idStr := c.Param("id")
	idUint, err := strconv.ParseUint(idStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	var req storageModel.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	data := req.GenData(uint16(idUint))
	if err := op.UpdateStorage(c.Request.Context(), &data); err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, data.GenResponse())
}

// @Summary 删除存储
// @Description 删除存储
// @Tags 存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "存储ID"
// @Success 200 {object} resp.ResponseStruct "删除成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/storage/{id} [delete]
func deleteStorage(c *gin.Context) {
	id := c.Param("id")
	idUint, err := strconv.ParseUint(id, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	if err := op.DeleteStorage(c.Request.Context(), uint16(idUint)); err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	resp.Success(c, nil)
}
