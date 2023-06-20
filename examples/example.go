package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	slogformatter "github.com/samber/slog-formatter"
	sloggin "github.com/samber/slog-gin"
	"golang.org/x/exp/slog"
)

func main() {
	// Create a slog logger, which:
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	logger := slog.New(
		slogformatter.NewFormatterHandler(
			slogformatter.TimezoneConverter(time.UTC),
			slogformatter.TimeFormatter(time.RFC3339, nil),
		)(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
		),
	)

	// Add an attribute to all log entries made through this logger.
	logger = logger.With("gin_mode", gin.EnvGinMode)

	router := gin.New()

	// Add the sloggin middleware to all routes.
	// The middleware will log all requests attributes under a "http" group.
	router.Use(sloggin.New(logger))

	// Example pong request.
	router.GET("/pong", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	logger.Info("Starting server")
	if err := router.Run(":1234"); err != nil {
		logger.Error("can' start server with 1234 port")
	}

	// output:
	// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="Incoming request" gin_mode=GIN_MODE http.status=200 http.method=GET http.path=/pong http.ip=127.0.0.1 http.latency=25.5µs http.user-agent=curl/7.77.0 http.time=2023-04-10T14:00:00.000+00:00
}
