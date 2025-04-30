package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	slogformatter "github.com/samber/slog-formatter"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracerProvider() (*sdktrace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resource, err := resource.New(ctx, resource.WithAttributes(
		attribute.String("service.name", "example"),
		attribute.String("service.namespace", "default"),
	))
	if err != nil {
		return nil, err
	}

	// tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(tp)

	// Set up a text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func main() {
	tp, _ := initTracerProvider()
	defer tp.Shutdown(context.Background())

	// Create a slog logger, which:
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	logger := slog.New(
		slogformatter.NewFormatterHandler(
			slogformatter.TimezoneConverter(time.UTC),
			slogformatter.TimeFormatter(time.RFC3339, nil),
		)(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}),
		),
	)

	// Add an attribute to all log entries made through this logger.
	logger = logger.With("gin_mode", gin.EnvGinMode)

	router := gin.New()
	router.Use(otelgin.Middleware("example"))

	// Add the sloggin middleware to all routes.
	// The middleware will log all requests attributes under a "http" group.
	//router.Use(sloggin.New(logger))
	config := sloggin.Config{
		WithRequestID: true,
		WithSpanID:    true,
		WithTraceID:   true,
	}
	router.Use(sloggin.NewWithConfig(logger, config))

	router.Use(gin.Recovery())

	// Example pong request.
	router.GET("/pong", func(c *gin.Context) {
		sloggin.AddCustomAttributes(c, slog.String("foo", "bar"))
		c.String(http.StatusOK, "pong")
	})
	router.GET("/pong/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.String(http.StatusOK, "pong %s", id)
	})

	logger.Info("Starting server")
	if err := router.Run(":4242"); err != nil {
		logger.Error("can' start server with 4242 port")
	}

	// output:
	// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="Incoming request" gin_mode=GIN_MODE http.status=200 http.method=GET http.path=/pong http.ip=127.0.0.1 http.latency=25.5µs http.user-agent=curl/7.77.0 http.time=2023-04-10T14:00:00.000+00:00
}
