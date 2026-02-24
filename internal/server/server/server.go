// Package server 提供 NodeHub 应用程序的入口点。
//
// @title NodeHub API
// @version 1.0.0
// @description NodeHub -  API 文档
// @description
// @description 这是 NodeHub 的 API 文档
// @description
// @description ## 认证
// @description 大多数接口需要使用 JWT 令牌进行认证。
// @description 认证时，请在 Authorization 头中包含 JWT 令牌：
// @description `Authorization: Bearer <your-jwt-token>`
// @description
// @description ## 错误响应
// @description 所有错误响应都遵循统一格式，包含 code、message 和 error 字段。
// @description
// @description ## 成功响应
// @description 所有成功响应都遵循统一格式，包含 code、message 和 data 字段。
//
// @contact.name NodeHub API 支持
// @contact.email support@nodehub.com
//
// @license.name GPL-3.0
// @license.url https://opensource.org/license/gpl-3-0
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 类型为 "Bearer"，后跟空格和 JWT 令牌。
//
// @tag.name 认证
// @tag.description 用户认证相关接口
//
// @tag.name 系统
// @tag.description 系统状态和健康检查接口
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/evilCYH/NodeHub/internal/config"
	_ "github.com/evilCYH/NodeHub/internal/server/handlers"
	"github.com/evilCYH/NodeHub/internal/server/middleware"
	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/evilCYH/NodeHub/static"
	"github.com/gin-gonic/gin"
)

const (
	defaultReadTimeout     = 30 * time.Second
	defaultWriteTimeout    = 30 * time.Second
	defaultIdleTimeout     = 60 * time.Second
	defaultShutdownTimeout = 30 * time.Second
	defaultMaxHeaderBytes  = 1 << 20 // 1MB
)

var server *Server

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

func Initialize() error {

	r, routerErr := setRouter()
	if routerErr != nil {
		return fmt.Errorf("failed to set router: %w", routerErr)
	}

	server = &Server{
		httpServer: &http.Server{
			Addr:           fmt.Sprintf("%s:%d", config.Base().Server.Host, config.Base().Server.Port),
			Handler:        r,
			ReadTimeout:    defaultReadTimeout,
			WriteTimeout:   defaultWriteTimeout,
			IdleTimeout:    defaultIdleTimeout,
			MaxHeaderBytes: defaultMaxHeaderBytes,
		},
		router: r,
	}
	return nil
}

func Start() error {
	if server == nil {
		return fmt.Errorf("HTTP server not initialized, please call Initialize() first")
	}

	log.Infof("Starting HTTP server %s", server.httpServer.Addr)

	go func() {
		if err := server.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Failed to start HTTP server: %v", err)
		}
	}()

	return nil
}

func Close() error {
	if server == nil {
		return fmt.Errorf("HTTP server not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := server.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server force closed: %v", err)
		return fmt.Errorf("HTTP server force closed: %w", err)
	}

	log.Debug("HTTP server closed")
	return nil
}

func IsInitialized() bool {
	return server != nil
}

func setRouter() (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// r.Use(middleware.Logging())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())
	r.Use(middleware.StaticEmbed("/", static.StaticFS))

	if err := router.RegisterAll(r); err != nil {
		return nil, fmt.Errorf("failed to register routes: %w", err)
	}

	log.Debugf("successfully registered %d routes", router.GetRouterCount())
	return r, nil
}
