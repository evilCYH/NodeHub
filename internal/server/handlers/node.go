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
			router.NewRoute("/detail", router.GET).
				Handle(getNodeDetail),
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
				SubID:        n.Base.SubId,
				UniqueKey:    n.Base.UniqueKey,
				UniqueKeyStr: strconv.FormatUint(n.Base.UniqueKey, 10),
				Name:         meta.Name,
				Type:         meta.Type,
				Country:      "",
				AliveStatus:  0,
				InitStatus:   nodeModel.InitPassed,
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
					UniqueKeyStr:    strconv.FormatUint(record.Base.UniqueKey, 10),
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
				SubID:        fn.SubID,
				UniqueKey:    fn.UniqueKey,
				UniqueKeyStr: strconv.FormatUint(fn.UniqueKey, 10),
				Name:         fn.Name,
				Type:         fn.Type,
				Reason:       fn.Reason,
				Delay:        0,
				SpeedUp:      0,
				SpeedDown:    0,
				Risk:         0,
				AliveStatus:  0,
				Country:      "",
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

func parseNodeMeta(raw []byte, fallbackName, fallbackType string) (nodeMeta, map[string]any) {
	meta := nodeMeta{Name: fallbackName, Type: fallbackType}
	var cfg map[string]any
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return meta, map[string]any{}
	}
	if name, ok := cfg["name"].(string); ok && name != "" {
		meta.Name = name
	}
	if nodeType, ok := cfg["type"].(string); ok && nodeType != "" {
		meta.Type = nodeType
	}
	return meta, cfg
}

// getNodeDetail 获取节点详情
// @Summary 获取节点详情
// @Description 通过 sub_id 与 unique_key 获取单个节点详情（含原始配置）
// @Tags 节点
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sub_id query int true "订阅ID"
// @Param unique_key query int true "节点唯一键"
// @Param scope query string false "数据范围：registry|pool，默认 registry"
// @Success 200 {object} resp.ResponseStruct{data=node.DetailResponse} "获取成功"
// @Failure 400 {object} resp.ResponseStruct "请求参数错误"
// @Failure 401 {object} resp.ResponseStruct "未授权"
// @Failure 404 {object} resp.ResponseStruct "节点不存在"
// @Router /api/v1/node/detail [get]
func getNodeDetail(c *gin.Context) {
	subIDStr := strings.TrimSpace(c.Query("sub_id"))
	uniqueKeyStr := strings.TrimSpace(c.Query("unique_key"))
	if uniqueKeyStr == "" {
		uniqueKeyStr = strings.TrimSpace(c.Query("unique_key_str"))
	}
	scope := strings.TrimSpace(c.Query("scope"))
	if scope == "" {
		scope = "registry"
	}
	if subIDStr == "" || uniqueKeyStr == "" {
		resp.ErrorBadRequest(c)
		return
	}
	subIDRaw, err := strconv.ParseUint(subIDStr, 10, 16)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	uniqueKey, err := strconv.ParseUint(uniqueKeyStr, 10, 64)
	if err != nil {
		resp.ErrorBadRequest(c)
		return
	}
	subID := uint16(subIDRaw)

	if scope == "pool" {
		nodes := node.GetAll()
		for _, n := range nodes {
			if n.Base.SubId != subID || n.Base.UniqueKey != uniqueKey {
				continue
			}
			meta, raw := parseNodeMeta(n.Base.Raw, "", "")
			item := nodeModel.DetailResponse{
				Response: nodeModel.Response{
					SubID:        n.Base.SubId,
					UniqueKey:    n.Base.UniqueKey,
					UniqueKeyStr: strconv.FormatUint(n.Base.UniqueKey, 10),
					Name:         meta.Name,
					Type:         meta.Type,
					Country:      "",
					AliveStatus:  0,
					InitStatus:   nodeModel.InitPassed,
				},
				Raw: raw,
			}
			if n.Info != nil {
				item.Delay = n.Info.Delay.Average()
				item.SpeedUp = n.Info.SpeedUp.Average()
				item.SpeedDown = n.Info.SpeedDown.Average()
				item.Risk = n.Info.Risk
				item.AliveStatus = n.Info.AliveStatus
				item.Country = n.Info.Country
			}
			resp.Success(c, item)
			return
		}
		resp.Error(c, 404, "node not found")
		return
	}

	records := node.GetRegistryBySubId([]uint16{subID})
	for _, record := range records {
		if record.Base.UniqueKey != uniqueKey {
			continue
		}
		meta, raw := parseNodeMeta(record.Base.Raw, "", "")
		item := nodeModel.DetailResponse{
			Response: nodeModel.Response{
				SubID:           record.Base.SubId,
				UniqueKey:       record.Base.UniqueKey,
				UniqueKeyStr:    strconv.FormatUint(record.Base.UniqueKey, 10),
				Name:            meta.Name,
				Type:            meta.Type,
				Country:         "",
				AliveStatus:     0,
				InitStatus:      record.InitStatus,
				LastCheckAt:     record.LastCheckAt,
				LastCheckSource: record.LastCheckSource,
				LastFailReason:  record.LastFailReason,
			},
			Raw: raw,
		}
		if record.Info != nil {
			item.Delay = record.Info.Delay.Average()
			item.SpeedUp = record.Info.SpeedUp.Average()
			item.SpeedDown = record.Info.SpeedDown.Average()
			item.Risk = record.Info.Risk
			item.AliveStatus = record.Info.AliveStatus
			item.Country = record.Info.Country
		}
		resp.Success(c, item)
		return
	}
	resp.Error(c, 404, "node not found")
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
// @Param run_id query int false "运行ID（可选，传入后按 run_id 精确查询）"
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
	var runID uint64
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
	if runIDStr := strings.TrimSpace(c.Query("run_id")); runIDStr != "" {
		parsedRunID, err := strconv.ParseUint(runIDStr, 10, 64)
		if err != nil {
			resp.ErrorBadRequest(c)
			return
		}
		runID = parsedRunID
	}
	resp.Success(c, node.GetUpdateLog(uint16(parsedID), limit, runID))
}
