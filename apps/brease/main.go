package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/loopfz/gadgeto/tonic"
	stats "github.com/semihalev/gin-stats"
	"github.com/speakeasy-api/speakeasy-go-sdk"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"go.dot.industries/brease/api"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("No environment variables")
	}

	logger, _, flush := tracer()
	defer flush()

	app := newApp(logger)

	_ = endless.ListenAndServe(getenv("HOST", ":4400"), app)
}

func newApp(logger *zap.Logger) *fizz.Fizz {
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
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

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

	r.GET("/", index)
	r.GET("/stats", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, stats.Report())
	})

	bh := &api.BreaseHandler{}

	f := fizz.NewFromEngine(r)
	infos := &openapi.Info{
		Title:       "brease API",
		Description: `Business rule engine as a service API spec.`,
		Version:     "0.1.0",
	}
	f.Generator().SetSecuritySchemes(map[string]*openapi.SecuritySchemeOrRef{
		"apiToken": {
			SecurityScheme: &openapi.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	})
	f.GET("/openapi.json", nil, f.OpenAPI(infos, "json"))

	grp := f.Group("/:contextID", "contextID", "Rule domain context")
	r.Use(ApiKeyAuthMiddleware(logger))

	// API methods
	grp.GET("/rules", []fizz.OperationOption{
		fizz.ID("getAllRules"),
	}, tonic.Handler(bh.AllRules, 200))
	grp.POST("/rules/add", []fizz.OperationOption{
		fizz.ID("addRule"),
	}, tonic.Handler(bh.AddRule, 200))
	grp.PUT("/rules/:id", []fizz.OperationOption{
		fizz.ID("replaceRule"),
	}, tonic.Handler(bh.ReplaceRule, 200))
	grp.DELETE("/rules/:id", []fizz.OperationOption{
		fizz.ID("removeRule"),
	}, tonic.Handler(bh.RemoveRule, 200))
	grp.POST("/evaluate", []fizz.OperationOption{
		fizz.ID("evaluateRules"),
	}, tonic.Handler(bh.ExecuteRules, 200))

	return f
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
