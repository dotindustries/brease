package main

import (
	"net/http"
	"os"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/speakeasy-api/speakeasy-go-sdk"
	"go.dot.industries/brease/api"
	"go.uber.org/zap"
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
	router := gin.Default()
	// https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies
	_ = router.SetTrustedProxies(nil)

	speakeasyApiKey := getenv("SPEAKEASY_API_KEY", "")
	if speakeasyApiKey != "" {
		// Configure the Global SDK
		speakeasy.Configure(speakeasy.Config{
			APIKey:    speakeasyApiKey,
			ApiID:     "brease",
			VersionID: "0.1",
		})
		router.Use(speakeasy.GinMiddleware)
		logger.Info("Configured Speakeasy API layer")
	}

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	api := &api.BreaseHandler{}

	router.GET("/", index)
	router.GET("/:contextID/rules", api.AllRules)
	router.POST("/:contextID/rules/add", api.AddRule)
	router.PUT("/:contextID/rules/:id", api.ReplaceRule)
	router.DELETE("/:contextID/rules/:id", api.DeleteRule)
	router.POST("/:contextID/execute", api.ExecuteRules)

	return router
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
