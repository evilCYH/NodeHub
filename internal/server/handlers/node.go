package handlers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/evilCYH/NodeHub/internal/core/node"
	nodeModel "github.com/evilCYH/NodeHub/internal/models/node"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func init() {
	router.NewGroupRouter("/api/v1/node").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", router.GET).
				Handle(getNodes),
		).
		AddRoute(
			router.NewRoute("/log", router.GET).
				Handle(getNodeUpdateLog),
		)
}

type nodeMeta struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// getNodes 获取节点列表
// @Summary 获取节点列表
// @Description 获取节点列表，可选 sub_id 过滤
// @Tags 节点
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sub_id query string false "订阅ID，支持逗号分隔"
// @Param include_failed query bool false "是否包含初测失败节点"
// @Param scope query string false "数据范围：registry|pool"
// @Param status query string false "状态筛选：alive|dead|init_failed|all"
// @Success 200 {object} resp.ResponseStruct{data=[]node.Response} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Router /api/v1/node [get]
func getNodes(c *gin.Context) {
	subIDRaw := strings.TrimSpace(c.Query("sub_id"))
	includeFailed := strings.TrimSpace(c.Query("include_failed"))
	scope := strings.TrimSpace(c.Query("scope"))
	status := strings.TrimSpace(c.Query("status"))
	var nodes []nodeModel.Data
	var records []nodeModel.Record
	var subIDs []uint16
	if scope == "" {
		scope = "registry"
	}
	if status == "" {
		status = "all"
	}
	if subIDRaw != "" {
		ids, err := parseSubIDs(subIDRaw)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		subIDs = ids
		if scope == "pool" {
			nodes = *node.GetBySubId(ids)
		} else {
			records = node.GetRegistryBySubId(ids)
		}
	} else {
		if scope == "pool" {
			nodes = node.GetAll()
		} else {
			records = node.GetRegistryAll()
		}
	}

	respData := make([]nodeModel.Response, 0)
	if scope == "pool" {
		respData = make([]nodeModel.Response, 0, len(nodes))
		for _, n := range nodes {
			var meta nodeMeta
			_ = yaml.Unmarshal(n.Base.Raw, &meta)
			if status != "all" {
				alive := n.Info != nil && (n.Info.AliveStatus&nodeModel.Alive != 0)
				if status == "alive" && !alive {
					continue
				}
				if status == "dead" && alive {
					continue
				}
				if status == "init_failed" {
					continue
				}
			}
			item := nodeModel.Response{
				SubID:       n.Base.SubId,
				UniqueKey:   n.Base.UniqueKey,
				Name:        meta.Name,
				Type:        meta.Type,
				Country:     "",
				AliveStatus: 0,
				InitStatus:  nodeModel.InitPassed,
			}
			if n.Info != nil {
				item.Delay = n.Info.Delay.Average()
				item.SpeedUp = n.Info.SpeedUp.Average()
				item.SpeedDown = n.Info.SpeedDown.Average()
				item.Risk = n.Info.Risk
				item.AliveStatus = n.Info.AliveStatus
				item.Country = n.Info.Country
			}
			respData = append(respData, item)
		}
	} else {
		respData = make([]nodeModel.Response, 0, len(records))
		for _, record := range records {
			var meta nodeMeta
			_ = yaml.Unmarshal(record.Base.Raw, &meta)
			alive := record.Info != nil && (record.Info.AliveStatus&nodeModel.Alive != 0)
			if status == "alive" && !alive {
				continue
			}
			if status == "dead" && alive {
				continue
			}
			if status == "init_failed" && record.InitStatus != nodeModel.InitFailed {
				continue
			}
			if status == "all" || status == "init_failed" || status == "alive" || status == "dead" {
				item := nodeModel.Response{
					SubID:           record.Base.SubId,
					UniqueKey:       record.Base.UniqueKey,
					Name:            meta.Name,
					Type:            meta.Type,
					Country:         "",
					AliveStatus:     0,
					InitStatus:      record.InitStatus,
					LastCheckAt:     record.LastCheckAt,
					LastCheckSource: record.LastCheckSource,
					LastFailReason:  record.LastFailReason,
				}
				if record.Info != nil {
					item.Delay = record.Info.Delay.Average()
					item.SpeedUp = record.Info.SpeedUp.Average()
					item.SpeedDown = record.Info.SpeedDown.Average()
					item.Risk = record.Info.Risk
					item.AliveStatus = record.Info.AliveStatus
					item.Country = record.Info.Country
				}
				respData = append(respData, item)
			}
		}
	}

	if scope != "registry" && (includeFailed == "true" || includeFailed == "1") {
		if len(subIDs) == 0 {
			subIDs = node.GetSubIDsFromPool()
		}
		failedNodes := node.GetFailedBySubId(subIDs)
		// 按 sub_id + unique_key 去重，避免同一节点在池内与失败列表中重复出现
		type nodeKey struct {
			subID     uint16
			uniqueKey uint64
		}
		seen := make(map[nodeKey]struct{}, len(respData))
		for _, item := range respData {
			seen[nodeKey{subID: item.SubID, uniqueKey: item.UniqueKey}] = struct{}{}
		}
		for _, fn := range failedNodes {
			key := nodeKey{subID: fn.SubID, uniqueKey: fn.UniqueKey}
			if _, exists := seen[key]; exists {
				continue
			}
			respData = append(respData, nodeModel.Response{
				SubID:       fn.SubID,
				UniqueKey:   fn.UniqueKey,
				Name:        fn.Name,
				Type:        fn.Type,
				Reason:      fn.Reason,
				Delay:       0,
				SpeedUp:     0,
				SpeedDown:   0,
				Risk:        0,
				AliveStatus: 0,
				Country:     "",
			})
		}
	}

	resp.Success(c, respData)
}

func parseSubIDs(raw string) ([]uint16, error) {
	parts := strings.Split(raw, ",")
	ids := make([]uint16, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 16)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uint16(id))
	}
	if len(ids) == 0 {
		return nil, errors.New("empty sub_id")
	}
	return ids, nil
}

// getNodeUpdateLog 获取订阅节点更新日志
// @Summary 获取订阅节点更新日志
// @Description 获取订阅节点更新日志
// @Tags 节点
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sub_id query int true "订阅ID"
// @Param limit query int false "返回条数"
// @Success 200 {object} resp.ResponseStruct{data=node.UpdateLogResponse} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Router /api/v1/node/log [get]
func getNodeUpdateLog(c *gin.Context) {
	subIDStr := strings.TrimSpace(c.Query("sub_id"))
	if subIDStr == "" {
		resp.ErrorBadRequest(c)
		return
	}
	parsedID, err := strconv.ParseUint(subIDStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	limit := 5
	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		if parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	resp.Success(c, node.GetUpdateLog(uint16(parsedID), limit))
}
