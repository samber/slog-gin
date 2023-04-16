package sloggin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level
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
	})
}

func NewWithConfig(logger *slog.Logger, config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		attributes := []slog.Attr{
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
			slog.String("user-agent", c.Request.UserAgent()),
			slog.Time("time", end),
		}

		switch {
		case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ClientErrorLevel, c.Errors.String(), attributes...)
		case c.Writer.Status() >= http.StatusInternalServerError:
			logger.LogAttrs(context.Background(), config.ServerErrorLevel, c.Errors.String(), attributes...)
		default:
			logger.LogAttrs(context.Background(), config.DefaultLevel, "HTTP Request", attributes...)
		}
	}
}
