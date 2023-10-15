package sloggin

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDCtx = "slog-gin.request-id"

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID bool

	Filters []Filter
}

// New returns a gin.HandlerFunc (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func New(logger *slog.Logger) gin.HandlerFunc {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,

		Filters: []Filter{},
	})
}

// NewWithFilters returns a gin.HandlerFunc (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewWithFilters(logger *slog.Logger, filters ...Filter) gin.HandlerFunc {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithRequestID: true,

		Filters: filters,
	})
}

// NewWithConfig returns a gin.HandlerFunc (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		requestID := uuid.New().String()
		if config.WithRequestID {
			c.Set(requestIDCtx, requestID)
			c.Header("X-Request-ID", requestID)
		}

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		attributes := []slog.Attr{
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("route", c.FullPath()),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.Time("time", end),
		}

		if config.WithRequestID {
			attributes = append(attributes, slog.String("request-id", requestID))
		}

		for _, filter := range config.Filters {
			if !filter(c) {
				return
			}
		}

		switch {
		case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ClientErrorLevel, c.Errors.String(), attributes...)
		case c.Writer.Status() >= http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ServerErrorLevel, c.Errors.String(), attributes...)
		default:
			logger.LogAttrs(context.Background(), config.DefaultLevel, "Incoming request", attributes...)
		}
	}
}

// GetRequestID returns the request identifier
func GetRequestID(c *gin.Context) string {
	requestID, ok := c.Get(requestIDCtx)
	if !ok {
		return ""
	}

	if id, ok := requestID.(string); ok {
		return id
	}

	return ""
}
