package handlers

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/evilCYH/NodeHub/internal/database/op"
	"github.com/evilCYH/NodeHub/internal/models/setting"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 日志等级优先级常量
var logLevelPriority = map[string]int{
	"debug": 0,
	"info":  1,
	"warn":  2,
	"error": 3,
	"fatal": 4,
}

// 默认配置
const (
	WriteTimeout      = 5  // 10秒
	PingInterval      = 5  // 5秒
	MaxConnections    = 20 // 20个连接
	WriteBufferSize   = 1024
	ChannelBufferSize = 256
)

// LogFilter 日志过滤器
type LogFilter struct {
	nameFilter  string
	levelFilter string
}

// ShouldSend 检查是否应该发送日志
func (f *LogFilter) ShouldSend(logEntry log.LogEntry) bool {
	if f.nameFilter != "" && !strings.Contains(logEntry.Name, f.nameFilter) {
		return false
	}

	if f.levelFilter != "" && !shouldSendLogLevel(f.levelFilter, logEntry.Level) {
		return false
	}

	return true
}

// shouldSendLogLevel 检查日志等级是否应该发送
func shouldSendLogLevel(filterLevel, logLevel string) bool {
	filterPriority, filterExists := logLevelPriority[filterLevel]
	logPriority, logExists := logLevelPriority[logLevel]

	if !filterExists || !logExists {
		return true
	}

	return logPriority >= filterPriority
}

// wsHandler WebSocket处理器
type wsHandler struct {
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]*Client
	mu          sync.RWMutex
	clientCount int32
}

// Client WebSocket客户端信息
type Client struct {
	conn   *websocket.Conn
	filter LogFilter
	send   chan log.LogEntry
	mu     sync.RWMutex
}

// init 函数用于自动注册路由
func init() {
	wsHandler := newWSHandler()

	router.NewGroupRouter("/api/v1/ws").
		Use(middleware.WSAuth()).
		AddRoute(
			router.NewRoute("/logs", router.GET).
				Handle(wsHandler.handleLogWebSocket),
		)
}

// newWSHandler 创建WebSocket处理器
func newWSHandler() *wsHandler {

	h := &wsHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if utils.IsDebug() {
					return true
				}

				origin := r.Header.Get("Origin")

				if origin == "" {
					log.Debugf("WebSocket客户端连接: 没有Origin头")
					return false
				}
				allowList := op.GetSettingStr(setting.WS_ORIGIN_ALLOWLIST)
				if allowList == "" {
					log.Warnf("WebSocket客户端连接: Origin=%s 被拒绝 (未配置允许列表)", origin)
					return false
				}
				allowed := splitAllowList(allowList)
				if !originAllowed(origin, allowed) {
					log.Warnf("WebSocket客户端连接: Origin=%s 被拒绝", origin)
					return false
				}

				log.Debugf("WebSocket客户端连接: Origin=%s", origin)

				return true
			},
			WriteBufferSize: WriteBufferSize,
		},
		clients: make(map[*websocket.Conn]*Client),
	}
	go h.broadcastLogs()
	return h
}

func splitAllowList(value string) []string {
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func originAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, item := range allowed {
		if strings.EqualFold(origin, item) {
			return true
		}
	}
	return false
}

func (h *wsHandler) handleLogWebSocket(c *gin.Context) {
	if atomic.LoadInt32(&h.clientCount) >= MaxConnections {
		resp.Error(c, http.StatusTooManyRequests, "connection limit reached")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("WebSocket升级失败: %v", err)
		return
	}

	nameFilter := c.Query("name")
	levelFilter := c.Query("level")

	client := &Client{
		conn: conn,
		filter: LogFilter{
			nameFilter:  nameFilter,
			levelFilter: levelFilter,
		},
		send: make(chan log.LogEntry, ChannelBufferSize),
	}

	h.mu.Lock()
	h.clients[conn] = client
	atomic.AddInt32(&h.clientCount, 1)
	h.mu.Unlock()

	username, _ := c.Get("username")
	clientIP := c.ClientIP()
	log.Infof("WebSocket客户端连接: 用户=%s, IP=%s, 当前连接数=%d", username, clientIP, atomic.LoadInt32(&h.clientCount))

	go h.handleClient(client)
}

func (h *wsHandler) broadcastLogs() {
	logChannel := log.GetWSChannel()

	for logEntry := range logChannel {
		h.broadcastToClients(logEntry)
	}
}

func (h *wsHandler) broadcastToClients(logEntry log.LogEntry) {
	var clientsToRemove []*websocket.Conn

	for conn, client := range h.clients {
		if h.shouldSendLog(client, logEntry) {
			select {
			case client.send <- logEntry:
			default:
				clientsToRemove = append(clientsToRemove, conn)
				log.Warnf("WebSocket客户端发送缓冲区满，移除客户端: %v", conn.RemoteAddr())
			}
		}
	}

	if len(clientsToRemove) > 0 {
		h.mu.Lock()
		for _, conn := range clientsToRemove {
			if client, exists := h.clients[conn]; exists {
				close(client.send)
				delete(h.clients, conn)
				atomic.AddInt32(&h.clientCount, -1)
			}
		}
		h.mu.Unlock()

		if len(clientsToRemove) > 0 {
			log.Warnf("移除了 %d 个缓冲区满的WebSocket客户端", len(clientsToRemove))
		}
	}
}

func (h *wsHandler) shouldSendLog(client *Client, logEntry log.LogEntry) bool {
	client.mu.RLock()
	defer client.mu.RUnlock()

	return client.filter.ShouldSend(logEntry)
}

func (h *wsHandler) handleClient(client *Client) {
	defer func() {
		h.removeClient(client)
		client.conn.Close()
	}()

	ticker := time.NewTicker(time.Duration(PingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case logEntry := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(time.Duration(WriteTimeout) * time.Second))
			if err := client.conn.WriteJSON(logEntry); err != nil {
				log.Errorf("WebSocket发送消息失败: %v", err)
				return
			}
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(time.Duration(WriteTimeout) * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Debugf("WebSocket ping失败，断开连接: %v", err)
				return
			}
		}
	}
}

func (h *wsHandler) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[client.conn]; exists {
		delete(h.clients, client.conn)
		close(client.send)
		atomic.AddInt32(&h.clientCount, -1)
		log.Debugf("WebSocket客户端断开连接, 当前连接数=%d", atomic.LoadInt32(&h.clientCount))
	}
}
