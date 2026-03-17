package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/middleware"
)

// InitRootRouterGroup creates a Gin engine with recovery, CORS, metrics, tracing, and optional access logging.
func InitRootRouterGroup(_ context.Context, config controller.ServerConfig, appName string) (*gin.Engine, error) {
	if !config.Debug.Gin {
		gin.SetMode(gin.ReleaseMode)
	}

	corsConfig, err := controller.InitCORS(&config.CORS)
	if err != nil {
		return nil, fmt.Errorf("init CORS: %w", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors.New(corsConfig))
	router.Use(middleware.PrometheusMiddleware())
	router.Use(otelgin.Middleware(appName, otelgin.WithFilter(func(req *http.Request) bool {
		return req.URL.Path != "/"
	})))

	if config.Log.AccessLog {
		router.Use(sloggin.NewWithConfig(slog.Default(), sloggin.Config{ //nolint:exhaustruct
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,
			WithRequestBody:  config.Log.AccessLogRequestBody,
			WithResponseBody: config.Log.AccessLogResponseBody,
			Filters: []sloggin.Filter{
				func(c *gin.Context) bool {
					path := c.Request.URL.Path
					return path != "/"
				},
			},
		}))
	}

	if config.Debug.Wait {
		router.Use(middleware.NewWaitMiddleware(time.Second))
	}

	return router, nil
}
