package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/log").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", router.GET).
				Handle(getLogFileList),
		).
		AddRoute(
			router.NewRoute("/content", router.GET).
				Handle(getLogContent),
		)
}

// @Summary 获取日志列表
// @Description 获取日志列表
// @Tags 日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "日志文件路径"
// @Success 200 {object} resp.ResponseStruct{data=[]uint64} "获取成功"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/log/list [get]
func getLogFileList(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		resp.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}
	logFileList, err := log.GetLogFileList(path)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	resp.Success(c, logFileList)
}

// @Summary 获取日志内容
// @Description 获取日志内容
// @Tags 日志
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param path query string true "日志文件路径"
// @Param timestamp query uint64 true "日志文件时间戳"
// @Success 200 {object} resp.ResponseStruct{data=[]object{level=string,time=string,msg=string}} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "文件不存在"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/log/content [get]
func getLogContent(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		resp.Error(c, http.StatusBadRequest, "path parameter is required")
		return
	}
	timestampStr := c.Query("timestamp")

	if timestampStr == "" {
		resp.Error(c, http.StatusBadRequest, "timestamp parameter is required")
		return
	}

	timestamp, err := strconv.ParseUint(timestampStr, 10, 64)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, "invalid timestamp format")
		return
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	c.Status(http.StatusOK)
	w := c.Writer

	w.WriteString(`{"code":200,"message":"success","data":[`)
	w.Flush()

	err = log.StreamLogToHTTP(path, timestamp, w)
	if err != nil {
		w.WriteString(`],"error":"`)
		w.WriteString(strings.ReplaceAll(err.Error(), `"`, `\"`))
		w.WriteString(`"}`)
		w.Flush()
		return
	}

	w.WriteString(`]}`)
	w.Flush()
}
