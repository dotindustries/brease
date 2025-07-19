package trace

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

var Tracer = otel.Tracer("gin-server")

func SpanNameFormatter(c *gin.Context) string {
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return c.FullPath()
}
