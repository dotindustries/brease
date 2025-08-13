package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/grpcreflect"
	"connectrpc.com/vanguard"
	"github.com/arl/statsviz"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	openapi2 "go.dot.industries/brease/openapi"
	trace2 "go.dot.industries/brease/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"buf.build/gen/go/dot/brease/connectrpc/go/brease/auth/v1/authv1connect"
	"buf.build/gen/go/dot/brease/connectrpc/go/brease/context/v1/contextv1connect"
	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"go.dot.industries/brease/auditlog"
	"go.dot.industries/brease/auditlog/auditlogstore"
	"go.dot.industries/brease/storage/redis"

	"github.com/fvbock/endless"
	"github.com/gin-contrib/requestid"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
	"go.dot.industries/brease/api"
	"go.dot.industries/brease/auth"
	"go.dot.industries/brease/cache/memory"
	"go.dot.industries/brease/env"
	log2 "go.dot.industries/brease/log"
	"go.dot.industries/brease/storage"
	"go.dot.industries/brease/storage/buntdb"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	nrgin "github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func main() {
	err := env.LoadEnv()
	if err != nil {
		log.Println("WARN: No .env file")
	} else {
		if env.IsDebug() {
			env.PrintEnv()
		}
	}

	logger, _, flush := log2.Logger()
	defer flush()

	otelShutdown, err := trace2.SetupOTelSDK(context.Background(), logger)
	if err != nil {
		logger.Fatal("OTel SDK setup failed", zap.Error(err))
		return
	}

	db, err := setupStorage(logger)
	if err != nil {
		logger.Fatal("failed to initialize storage", zap.Error(err))
	}
	defer func() {
		dbErr := db.Close()
		err = errors.Join(err, dbErr, otelShutdown(context.Background()))
	}()

	app := newApp(db, logger)
	host := env.Getenv("HOST", "0.0.0.0")
	port := env.Getenv("PORT", "4400")
	addr := fmt.Sprintf("%s:%s", host, port)

	err = endless.ListenAndServe(addr, app)
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

func newRelicApm(logger *zap.Logger) (*newrelic.Application, error) {
	appName := env.Getenv("NEW_RELIC_APP_NAME", "")
	license := env.Getenv("NEW_RELIC_LICENSE", "")

	if appName == "" && license == "" {
		logger.Debug("Skipping New Relic setup: not required.")
		// not set up
		return nil, nil
	}
	return newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(license),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
}

func newApp(db storage.Database, logger *zap.Logger) *gin.Engine {
	if !env.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.UseH2C = true

	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies
	_ = r.SetTrustedProxies(nil)

	apm, err := newRelicApm(logger)
	if err != nil {
		logger.Fatal("Failed to set up New Relic APM", zap.Error(err))
		return nil
	}
	if apm != nil {
		r.Use(nrgin.Middleware(apm))
	}

	if os.Getenv("SENTRY_DSN") != "" {
		r.Use(sentrygin.New(sentrygin.Options{}))
		region := os.Getenv("FLY_REGION")
		if region != "" {
			r.Use(func(ctx *gin.Context) {
				if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
					hub.Scope().SetTag("region", region)
				}
				ctx.Next()
			})
		}
	}

	// config CORS
	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowMethods = append(connectcors.AllowedMethods(), "OPTIONS")
	config.AllowHeaders = connectcors.AllowedHeaders()
	config.ExposeHeaders = connectcors.ExposedHeaders()
	var allowedOrigins []string
	originsStr := env.Getenv("BREASE_CORS_ALLOW_ORIGINS", "*")
	if originsStr == "*" {
		config.AllowAllOrigins = true
		allowedOrigins = append(allowedOrigins, "*")
	} else {
		origins := strings.Split(originsStr, ",")
		config.AllowOrigins = append(config.AllowOrigins, origins...)
		allowedOrigins = append(allowedOrigins, origins...)
		logger.Info(
			"CORS origins",
			zap.Strings("origins", origins),
		)
		config.AllowOriginFunc = func(origin string) bool {
			slog.Error("Trying origin check", "origin", origin, "allowedOrigin", origins)
			return true
		}
	}

	r.Use(cors.New(config))
	r.Use(requestid.New())
	r.Use(otelgin.Middleware(otelServiceName(), otelgin.WithSpanNameFormatter(trace2.SpanNameFormatter)))
	r.Use(stats.RequestStats())
	r.Use(ginzap.GinzapWithConfig(logger, ginLoggerConfig()))
	oaAssets, err := static.EmbedFolder(openapi2.OpenApiAssets, "assets")
	if err != nil {
		panic(err)
	}
	r.Use(static.Serve("/", oaAssets))
	r.Use(gin.Recovery())
	//r.Use(auth.Middleware(logger, []*regexp.Regexp{regexp.MustCompile("^/(brease.*|v1.*)$")}))
	r.Use(auditlog.Middleware(
		auditLogStore(logger),
		auditlog.WithSensitivePaths([]*regexp.Regexp{regexp.MustCompile("^/(token|refreshToken)$")}),
		auditlog.WithIgnorePaths([]*regexp.Regexp{regexp.MustCompile("^/(stats)$")}),
		auditlog.WithIDExtractor(func(c *gin.Context) (contextID, ownerID, userID string) {
			ownerID = c.GetString(auth.ContextOrgKey)
			// TODO: we don't have access to the contextID yet
			contextID = ""
			userID = c.GetString(auth.ContextUserIDKey)
			if userID == "" {
				userID = "n/a:" + ownerID
			}
			return
		}),
	))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"client": c.ClientIP(),
			"status": "ready to rumble!",
		})
	})
	r.GET("/stats", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, stats.Report())
	})
	// region setup server stats
	srv, _ := statsviz.NewServer()
	statsWs := srv.Ws()
	statsIndex := srv.Index()
	r.GET("/debug/statsviz/*filepath", func(context *gin.Context) {
		if context.Param("filepath") == "/ws" {
			statsWs(context.Writer, context.Request)
			return
		}
		statsIndex(context.Writer, context.Request)
	})
	// endregion

	// health
	checker := grpchealth.NewStaticChecker(
		authv1connect.AuthServiceName,
		contextv1connect.ContextServiceName,
	)
	healthPath, healthHandler := grpchealth.NewHandler(checker)
	r.GET(healthPath+"/*path", gin.WrapH(healthHandler))

	bh := api.NewHandler(db, memory.New(), logger)

	// connect auth
	authPath, authHandler := authv1connect.NewAuthServiceHandler(bh, connect.WithInterceptors(auth.NewAuthInterceptor(logger)))
	authService := vanguard.NewService(authPath, authHandler)
	r.Match([]string{"POST"}, authPath+"/*path", gin.WrapH(authHandler))

	// connect context
	ctxPath, ctxHandler := contextv1connect.NewContextServiceHandler(bh, connect.WithInterceptors(auth.NewAuthInterceptor(logger)))
	ctxService := vanguard.NewService(ctxPath, ctxHandler)
	r.Match([]string{"POST", "GET", "PATCH", "DELETE"}, ctxPath+"/*path", gin.WrapH(ctxHandler))

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
		authService,
		ctxService,
	})
	if err != nil {
		logger.Fatal("Failed to set up vanguard transcoder")
	}

	mux := http.NewServeMux()
	mux.Handle("/", withCORS(transcoder, allowedOrigins))

	// add grpc reflection support for tools like `buf curl` or `grpcurl`
	mux.Handle(grpcreflect.NewHandlerV1(grpcreflect.NewStaticReflector(authv1connect.AuthServiceName, contextv1connect.ContextServiceName)))

	// register openapi handlers
	r.Any("/v1/*any", gin.WrapF(mux.ServeHTTP))

	return r
}

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(h http.Handler, origins []string) http.Handler {
	middleware := cors2.New(cors2.Options{
		AllowedOrigins:   origins,
		AllowCredentials: true,
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowOriginFunc: func(origin string) bool {
			slog.Info("Trying origin check", "origin", origin, "allowedOrigin", origins)
			return true
		},
	})
	return middleware.Handler(h)
}

func auditLogStore(logger *zap.Logger) auditlog.Store {
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

// FIXME: remove this -- missing features and oas3 support :/
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
	return env.Getenv("OTEL_SERVICE_NAME", "brease")
}
