package main

import (
	"bytes"
	"context"
	"fmt"
	"go.dot.industries/brease/auditlog"
	"go.dot.industries/brease/auditlog/auditlogstore"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"

	"go.dot.industries/brease/storage/redis"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

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
	"go.dot.industries/brease/cache/memory"
	"go.dot.industries/brease/env"
	log2 "go.dot.industries/brease/log"
	"go.dot.industries/brease/storage"
	"go.dot.industries/brease/storage/buntdb"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	err := env.LoadEnv()
	if err != nil {
		log.Println("WARN: No .env file")
	}

	if env.IsDebug() {
		env.PrintEnv()
	}

	logger, _, flush := log2.Logger()
	defer flush()

	cleanup := tracer(logger)
	defer cleanup(context.Background())

	db, err := setupStorage(logger)
	if err != nil {
		logger.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer db.Close()

	app := newApp(db, logger)

	host := env.Getenv("HOST", "")
	port := env.Getenv("PORT", "4400")
	addr := fmt.Sprintf("%s:%s", host, port)
	_ = endless.ListenAndServe(addr, app)
}

// setupStorage Determines which storage engine should be instantiated and returns an instance.
func setupStorage(logger *zap.Logger) (db storage.Database, err error) {
	if redisURL := env.Getenv("REDIS_URL", ""); redisURL != "" {
		return redis.NewDatabase(redis.Options{
			URL:    redisURL,
			Logger: logger,
		})
	}

	// memory db as fallback
	return buntdb.NewDatabase(buntdb.Options{Logger: logger})
}

func newApp(db storage.Database, logger *zap.Logger) *fizz.Fizz {
	if !env.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies
	_ = r.SetTrustedProxies(nil)

	r.Use(otelgin.Middleware(otelServiceName()))
	r.Use(requestid.New())
	r.Use(stats.RequestStats())
	r.Use(ginzap.GinzapWithConfig(logger, ginLoggerConfig()))

	r.Use(auditlog.Middleware(
		getAuditLogStore(logger),
		auditlog.WithSensitivePaths([]*regexp.Regexp{regexp.MustCompile("^/(token|refreshToken)$")}),
		auditlog.WithIDExtractor(func(c *gin.Context) (contextID, ownerID, userID string) {
			ownerID = c.GetString(auth.ContextOrgKey)
			// TODO: we don't have access to the contextID yet
			contextID = ""
			userID = c.GetString(auth.ContextUserIDKey)
			if userID == "" {
				userID = "root:" + ownerID
			}
			return
		}),
	))
	r.Use(gin.Recovery())

	speakeasyAPIKey := env.Getenv("SPEAKEASY_API_KEY", "")
	if speakeasyAPIKey != "" {
		auth.InitJWKS()

		// Configure the Global SDK
		speakeasy.Configure(speakeasy.Config{
			APIKey:    speakeasyAPIKey,
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

	f := newOpenapi(r)
	f.GET("/openapi.json", nil, api.OpenAPISpecHandler(f, logger))

	bh := api.NewHandler(db, memory.New(), logger)
	tonic.SetErrorHook(jujerr.ErrHook)

	security := &openapi.SecurityRequirement{
		"JWTAuth":    []string{},
		"ApiKeyAuth": []string{},
	}

	authGrp := f.Group("/", "auth", "Authentication")
	authGrp.POST("/token", []fizz.OperationOption{
		fizz.ID("getToken"),
		fizz.Description("Generate a short lived access token for web access"),
		fizz.Security(security),
	}, auth.AuthMiddleware(logger), tonic.Handler(bh.GenerateTokenPair, 200))
	authGrp.POST("/refreshToken", []fizz.OperationOption{
		fizz.ID("refreshToken"),
		fizz.Description("Refresh the short lived access token for web access"),
	}, tonic.Handler(bh.RefreshTokenPair, 200))

	grp := f.Group("/:contextID", "context", "Ruleset domain context")
	grp.Use(auth.AuthMiddleware(logger))

	grp.GET("/rules", []fizz.OperationOption{
		fizz.ID("getAllRules"),
		fizz.Description("Returns all rules with the context"),
		fizz.Security(security),
	}, tonic.Handler(bh.AllRules, 200))
	grp.GET("/rules/:id/versions", []fizz.OperationOption{
		fizz.ID("getRuleVersions"),
		fizz.Description("Returns all versions of a rule"),
		fizz.Security(security),
	}, tonic.Handler(bh.GetRuleVersions, 200))
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

func getAuditLogStore(logger *zap.Logger) auditlog.Store {
	stores := auditlog.Stores{auditlogstore.NewLog(auditlogstore.LogConfig{Verbosity: 5}, logger)}
	if redisURL := env.Getenv("REDIS_URL", ""); redisURL != "" {
		redisStore, err := auditlogstore.NewRedis(auditlogstore.Options{
			URL:    redisURL,
			Logger: logger,
		})
		if err != nil {
			panic(err)
		}
		return append(stores, redisStore)
	}

	return stores
}

func ginLoggerConfig() *ginzap.Config {
	return &ginzap.Config{
		UTC:        true,
		TimeFormat: time.RFC3339,
		Context: func(c *gin.Context) []zapcore.Field {
			var fields []zapcore.Field
			// log request ID
			if requestID := c.Writer.Header().Get("X-Request-ID"); requestID != "" {
				fields = append(fields, zap.String("request_id", requestID))
			}

			// log trace and span ID
			if spanCtx := trace.SpanFromContext(c.Request.Context()).SpanContext(); spanCtx.IsValid() {
				fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()))
				fields = append(fields, zap.String("span_id", spanCtx.SpanID().String()))
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
	}
}

func newOpenapi(r *gin.Engine) *fizz.Fizz {
	f := fizz.NewFromEngine(r)
	f.Generator().SetInfo(&openapi.Info{
		Title:       "brease API",
		Description: `Business rule engine as a service`,
		Version:     "0.1.0",
		Contact: &openapi.Contact{
			Name:  "Support",
			URL:   "https://app.brease.run/support",
			Email: "support@dot.industries",
		},
		License: &openapi.License{
			Name: "MIT License",
			URL:  "https://opensource.org/licenses/MIT",
		},
	})
	f.Generator().SetServers([]*openapi.Server{
		{URL: "https://api.brease.run", Description: "Cloud hosted production server"},
		{URL: "http://localhost:4400", Description: "Development server"},
	})
	f.Generator().SetSecuritySchemes(map[string]*openapi.SecuritySchemeOrRef{
		"JWTAuth": {
			SecurityScheme: &openapi.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "Example header: \n> Authorization: JWT <token>",
			},
		},
	})
	return f
}

func otelServiceName() string {
	return env.Getenv("OTEL_SERVICE_NAME", "")
}

func tracer(logger *zap.Logger) func(context.Context) error {
	serviceName := otelServiceName()
	serviceVersion := env.Getenv("OTEL_VERSION", "0.0.1")
	serviceEnv := env.Getenv("OTEL_ENV", "dev")
	otelCollectorURL := env.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	otelInsecure := env.Getenv("OTEL_INSECURE_MODE", "")
	accessToken := env.Getenv("OTEL_ACCESS_TOKEN", "")

	if serviceName == "" {
		serviceName = "brease"
	}

	var exporter sdktrace.SpanExporter
	if otelCollectorURL == "" {
		var err error
		var opts []stdouttrace.Option
		if pretty := env.Getenv("OTEL_PRETTY_PRINT", ""); pretty != "" {
			opts = append(opts, stdouttrace.WithPrettyPrint())
		}
		exporter, err = stdouttrace.New(opts...)
		if err != nil {
			logger.Fatal("Failed to setup OTLP tracer", zap.Error(err))
		} else {
			logger.Debug("OTLP setup finished with stdout tracer")
		}
	} else {
		securityOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
		if len(otelInsecure) > 0 {
			securityOption = otlptracegrpc.WithInsecure()
		}
		var err error
		opts := []otlptracegrpc.Option{
			securityOption,
			otlptracegrpc.WithEndpoint(otelCollectorURL),
		}

		if accessToken != "" {
			opts = append(opts, otlptracegrpc.WithHeaders(map[string]string{
				"lightstep-access-token": accessToken,
			}))
		}

		exporter, err = otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(opts...),
		)
		if err != nil {
			logger.Fatal("failed to connect to OTLP tracer", zap.Error(err))
		} else {
			logger.Debug("OTLP setup finished with remote tracer", zap.String("collectorURL", otelCollectorURL))
		}
	}

	resources, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("library.language", "go"),
			attribute.String("environment", serviceEnv),
		),
	)
	if err != nil {
		logger.Fatal("Could not set resources: ", zap.Error(err))
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)
	return tp.Shutdown
}
