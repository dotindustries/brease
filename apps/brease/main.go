package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/loopfz/gadgeto/tonic/utils/jujerr"
	stats "github.com/semihalev/gin-stats"
	"github.com/speakeasy-api/speakeasy-go-sdk"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"go.dot.industries/brease/api"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/env"
	log2 "go.dot.industries/brease/log"
	"go.dot.industries/brease/storage"
	"go.dot.industries/brease/storage/buntdb"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	err := env.LoadEnv()

	if err != nil {
		log.Fatal("No environment variables")
	}

	logger, _, flush := log2.Logger()
	defer flush()

	db := setupStorage(logger)
	defer db.Close()

	app := newApp(db, logger)

	_ = endless.ListenAndServe(env.Getenv("HOST", ":4400"), app)
}

// setupStorage Determines which storage engine should be instantiated and returns an instance.
func setupStorage(logger *zap.Logger) storage.Database {
	db, err := buntdb.NewDatabase(buntdb.BuntDbOptions{Logger: logger})
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}
	return db
}

func newApp(db storage.Database, logger *zap.Logger) *fizz.Fizz {
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
	r.Use(gin.Recovery())

	speakeasyApiKey := env.Getenv("SPEAKEASY_API_KEY", "")
	if speakeasyApiKey != "" {
		auth.InitJWKS()

		// Configure the Global SDK
		speakeasy.Configure(speakeasy.Config{
			APIKey:    speakeasyApiKey,
			ApiID:     "brease",
			VersionID: "0.1",
		})
		r.Use(speakeasy.GinMiddleware)
		logger.Info("Configured Speakeasy API layer")
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"client": c.ClientIP(),
			"status": "ready to rumble!",
		})
	})
	r.GET("/stats", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, stats.Report())
	})

	bh := api.NewHandler(db, logger)

	tonic.SetErrorHook(jujerr.ErrHook)
	f := fizz.NewFromEngine(r)
	infos := &openapi.Info{
		Title:       "brease API",
		Description: `Business rule engine as a service API spec.`,
		Version:     "0.1.0",
		Contact: &openapi.Contact{
			Name:  "Brease API Support",
			URL:   "https://app.brease.run/support",
			Email: "support@dot.industries",
		},
	}
	f.Generator().SetServers([]*openapi.Server{
		{URL: "http://localhost:4400", Description: "Development server"},
		{URL: "https://api.brease.run", Description: "Cloud hosted production server"},
	})
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
	grp.Use(auth.ApiKeyAuthMiddleware(logger))

	security := &openapi.SecurityRequirement{
		"apiToken": []string{},
	}
	// API methods
	grp.GET("/rules", []fizz.OperationOption{
		fizz.ID("getAllRules"),
		fizz.Description("Returns all rules with the context"),
		fizz.Security(security),
	}, tonic.Handler(bh.AllRules, 200))
	grp.POST("/rules/add", []fizz.OperationOption{
		fizz.ID("addRule"),
		fizz.Description("Adds a new rule to the context"),
		fizz.Security(security),
	}, tonic.Handler(bh.AddRule, 200))
	grp.PUT("/rules/:id", []fizz.OperationOption{
		fizz.ID("replaceRule"),
		fizz.Description("Replaces an existing rule within the context"),
		fizz.Security(security),
	}, tonic.Handler(bh.ReplaceRule, 200))
	grp.DELETE("/rules/:id", []fizz.OperationOption{
		fizz.ID("removeRule"),
		fizz.Description("Removes a rule from the context"),
		fizz.Security(security),
	}, tonic.Handler(bh.RemoveRule, 200))
	grp.POST("/evaluate", []fizz.OperationOption{
		fizz.ID("evaluateRules"),
		fizz.Description("Evaluate rules within a context on the provided object"),
		fizz.Security(security),
	}, tonic.Handler(bh.EvaluateRules, 200))

	return f
}
