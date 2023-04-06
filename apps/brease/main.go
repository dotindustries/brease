package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
	"github.com/speakeasy-api/speakeasy-go-sdk"
	"go.dot.industries/brease/api"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func main() {
	var flush func()
	logger, _, flush = tracer()
	defer flush()

	app := newApp()

	err := endless.ListenAndServe(getenv("HOST", ":4400"), app)
	if err != nil {
		panic(err)
	}
}

func newApp() *gin.Engine {
	r := gin.Default()
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies
	_ = r.SetTrustedProxies(nil)

	r.Use(requestid.New())
	r.Use(stats.RequestStats())
	r.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: func(c *gin.Context) []zapcore.Field {
			var fields []zapcore.Field
			// log request ID
			if requestID := c.Writer.Header().Get("X-Request-ID"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
				fields = append(fields, zap.String("trace_id", trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
				fields = append(fields, zap.String("span_id", trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
			}

			// log request body
			var body []byte
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ = io.ReadAll(tee)
			c.Request.Body = io.NopCloser(&buf)
			fields = append(fields, zap.String("body", string(body)))

			return fields
		},
	}))

	speakeasyApiKey := getenv("SPEAKEASY_API_KEY", "")
	if speakeasyApiKey != "" {
		// Configure the Global SDK
		speakeasy.Configure(speakeasy.Config{
			APIKey:    speakeasyApiKey,
			ApiID:     "brease",
			VersionID: "0.1",
		})
		r.Use(speakeasy.GinMiddleware)
		logger.Info("Configured Speakeasy API layer")
	}

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	bh := &api.BreaseHandler{}

	r.GET("/", index)
	r.GET("/stats", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, stats.Report())
	})
	r.GET("/:contextID/rules", bh.AllRules)
	r.POST("/:contextID/rules/add", bh.AddRule)
	r.PUT("/:contextID/rules/:id", bh.ReplaceRule)
	r.DELETE("/:contextID/rules/:id", bh.DeleteRule)
	r.POST("/:contextID/execute", bh.ExecuteRules)

	return r
}

func index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"client": c.ClientIP(),
		"status": "ready to rumble!",
	})
}

func getenv(key string, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}
