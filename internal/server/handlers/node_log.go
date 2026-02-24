package handlers

import (
	"net/http"
	"strconv"

	"github.com/evilCYH/NodeHub/internal/database/op"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/node/log/detail").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(GetNodeTestLogs),
		)

	router.NewGroupRouter("/api/v1/node/logs").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(GetNodeLogs),
		)
}

// GetNodeTestLogs godoc
// @Summary 获取节点详细测试日志
// @Description 获取订阅下各节点的详细测试日志，支持分页、级别筛选和关键词搜索
// @Tags nodes
// @Param sub_id query int true "订阅ID"
// @Param level query string false "级别筛选: info/warn/error"
// @Param keyword query string false "关键词搜索（节点名或日志内容）"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认50，最大100" default(50)
// @Success 200 {object} map[string]any{code=int,data=node.NodeTestLogResponse}
// @Router /api/v1/node/log/detail [get]
func GetNodeTestLogs(c *gin.Context) {
	var query nodeModel.NodeTestLogQuery

	subID, err := strconv.Atoi(c.Query("sub_id"))
	if err != nil || subID <= 0 {
		resp.Error(c, http.StatusBadRequest, "invalid sub_id")
		return
	}
	query.SubID = uint16(subID)

	query.Level = c.DefaultQuery("level", "")
	query.Keyword = c.DefaultQuery("keyword", "")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	query.Page = page

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	query.PageSize = pageSize

	logs, total, err := op.QueryNodeLogs(c.Request.Context(), nodeModel.NodeLogQuery{
		SubID:    query.SubID,
		Source:   string(nodeModel.LogSourceInit),
		Level:    query.Level,
		Keyword:  query.Keyword,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	result := make([]nodeModel.NodeTestLog, 0, len(logs))
	for _, entry := range logs {
		result = append(result, nodeModel.NodeTestLog{
			ID:        entry.ID,
			SubID:     entry.SubID,
			NodeName:  entry.NodeName,
			Level:     entry.Level,
			Message:   entry.Message,
			CreatedAt: entry.CreatedAt,
		})
	}

	resp.Success(c, nodeModel.NodeTestLogResponse{
		Total: total,
		List:  result,
	})
}

// GetNodeLogs godoc
// @Summary 获取节点日志（初测+检测）
// @Description 获取订阅下节点日志，支持来源/级别/关键词/检测ID筛选
// @Tags nodes
// @Param sub_id query int true "订阅ID"
// @Param source query string false "来源筛选: init/check"
// @Param level query string false "级别筛选: info/warn/error"
// @Param keyword query string false "关键词搜索（节点名或日志内容）"
// @Param check_id query int false "检测任务ID"
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认50，最大200" default(50)
// @Success 200 {object} map[string]any{code=int,data=node.NodeLogResponse}
// @Router /api/v1/node/logs [get]
func GetNodeLogs(c *gin.Context) {
	var query nodeModel.NodeLogQuery

	subID, err := strconv.Atoi(c.Query("sub_id"))
	if err != nil || subID <= 0 {
		resp.Error(c, http.StatusBadRequest, "invalid sub_id")
		return
	}
	query.SubID = uint16(subID)
	query.Source = c.DefaultQuery("source", "")
	query.Level = c.DefaultQuery("level", "")
	query.Keyword = c.DefaultQuery("keyword", "")

	if checkIDStr := c.Query("check_id"); checkIDStr != "" {
		if checkID, err := strconv.Atoi(checkIDStr); err == nil && checkID > 0 {
			query.CheckID = uint16(checkID)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	query.Page = page

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	query.PageSize = pageSize

	logs, total, err := op.QueryNodeLogs(c.Request.Context(), query)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp.Success(c, nodeModel.NodeLogResponse{
		Total: total,
		List:  logs,
	})
}
