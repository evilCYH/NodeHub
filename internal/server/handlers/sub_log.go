package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bestruirui/bestsub/internal/database/op"
	subModel "github.com/bestruirui/bestsub/internal/models/sub"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/gin-gonic/gin"
)

// route registered in sub.go

// getSubRunLog 获取订阅运行日志
// @Summary 获取订阅运行日志
// @Description 获取订阅拉取/转换日志，包含运行记录与事件
// @Tags 订阅
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sub_id query int true "订阅ID"
// @Param limit query int false "返回条数" default(10)
// @Param include_events query bool false "是否包含历史 events" default(false)
// @Success 200 {object} resp.ResponseStruct{data=sub.RunLogResponse} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 500 {object} resp.ResponseStruct "服务器内部错误"
// @Router /api/v1/sub/log [get]
func getSubRunLog(c *gin.Context) {
	subIDStr := c.Query("sub_id")
	if subIDStr == "" {
		resp.ErrorBadRequest(c)
		return
	}
	parsedID, err := strconv.ParseUint(subIDStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		if parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	runs, err := op.ListSubRuns(c.Request.Context(), uint16(parsedID), limit)
	if err != nil {
		resp.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	includeEvents := c.DefaultQuery("include_events", "false")
	var events []subModel.RunEvent
	if len(runs) > 0 {
		if includeEvents == "true" || includeEvents == "1" {
			runTimeByID := make(map[uint64]time.Time, len(runs))
			for _, run := range runs {
				runTimeByID[run.ID] = run.CreatedAt
			}
			for _, run := range runs {
				runEvents, err := op.ListSubRunEvents(c.Request.Context(), run.ID, uint16(parsedID))
				if err != nil {
					resp.Error(c, http.StatusInternalServerError, err.Error())
					return
				}
				for i := range runEvents {
					runEvents[i].RunTime = runTimeByID[runEvents[i].RunID]
				}
				events = append(events, runEvents...)
			}
		} else {
			events, err = op.ListSubRunEvents(c.Request.Context(), runs[0].ID, uint16(parsedID))
			if err != nil {
				resp.Error(c, http.StatusInternalServerError, err.Error())
				return
			}
			for i := range events {
				events[i].RunTime = runs[0].CreatedAt
			}
		}
	}

	resp.Success(c, subModel.RunLogResponse{
		Runs:   runs,
		Events: events,
	})
}
