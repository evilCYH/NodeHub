//go:build debug

package handlers

import (
	"net/http/pprof"

	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {

	router.NewGroupRouter("/debug/pprof").
		AddRoute(
			router.NewRoute("/", router.GET).
				Handle(index),
		).
		AddRoute(
			router.NewRoute("/cmdline", router.GET).
				Handle(cmdline),
		).
		AddRoute(
			router.NewRoute("/profile", router.GET).
				Handle(profile),
		).
		AddRoute(
			router.NewRoute("/symbol", router.GET).
				Handle(symbol),
		).
		AddRoute(
			router.NewRoute("/symbol", router.POST).
				Handle(symbol),
		).
		AddRoute(
			router.NewRoute("/trace", router.GET).
				Handle(trace),
		).
		AddRoute(
			router.NewRoute("/allocs", router.GET).
				Handle(allocs),
		).
		AddRoute(
			router.NewRoute("/block", router.GET).
				Handle(block),
		).
		AddRoute(
			router.NewRoute("/goroutine", router.GET).
				Handle(goroutine),
		).
		AddRoute(
			router.NewRoute("/heap", router.GET).
				Handle(heap),
		).
		AddRoute(
			router.NewRoute("/mutex", router.GET).
				Handle(mutex),
		).
		AddRoute(
			router.NewRoute("/threadcreate", router.GET).
				Handle(threadcreate),
		)
}

func index(c *gin.Context) {
	pprof.Index(c.Writer, c.Request)
}

func cmdline(c *gin.Context) {
	pprof.Cmdline(c.Writer, c.Request)
}

func profile(c *gin.Context) {
	pprof.Profile(c.Writer, c.Request)
}

func symbol(c *gin.Context) {
	pprof.Symbol(c.Writer, c.Request)
}

func trace(c *gin.Context) {
	pprof.Trace(c.Writer, c.Request)
}

func allocs(c *gin.Context) {
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

func block(c *gin.Context) {
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

func goroutine(c *gin.Context) {
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

func heap(c *gin.Context) {
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

func mutex(c *gin.Context) {
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

func threadcreate(c *gin.Context) {
	pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
}
