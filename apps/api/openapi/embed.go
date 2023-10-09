package openapi

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
)

//go:embed all:assets
var assets embed.FS

// SetOpenAPIUIRoutes  sets api routes
func SetOpenAPIUIRoutes(e *gin.Engine) error {
	a, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	r := e.Group("/openapi/")
	r.StaticFS("/", http.FS(a))
	e.GET("/openapi", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/openapi/")
	})

	return nil
}
