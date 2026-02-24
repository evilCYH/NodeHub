//go:build dev

package handlers

import (
	"net/http"

	"github.com/evilCYH/NodeHub/internal/server/router"
	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/scalar").
		AddRoute(
			router.NewRoute("/", router.GET).
				Handle(scalar),
		).
		AddRoute(
			router.NewRoute("/api.json", router.GET).
				Handle(apidata),
		)
}

var scalarHTML = []byte(`
<!doctype html>
<html>
  <head>
    <title>NodeHub API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>Scalar.createApiReference(
    '#app', 
      {
        url: '/scalar/api.json',
        hideModels: true,
        hideDownloadButton: true,
        authentication: {
          preferredSecurityScheme: 'BearerAuth',
        },
        hideClientButton: true
      }
    )
    </script>
  </body>
</html>
`)

func scalar(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", scalarHTML)
}
func apidata(c *gin.Context) {
	c.File("docs/api/swagger.json")
}
