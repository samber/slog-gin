package sloggin

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	customAttributesCtxKey = "slog-gin.custom-attributes"
	requestIDCtx           = "slog-gin.request-id"
)

var (
	RequestBodyMaxSize  = 64 * 1024 // 64KB
	ResponseBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
		"authorization": {},
		"cookie":        {},
		"set-cookie":    {},
		"x-auth-token":  {},
		"x-csrf-token":  {},
		"x-xsrf-token":  {},
	}
	HiddenResponseHeaders = map[string]struct{}{
		"set-cookie": {},
	}
)

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithRequestID      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool

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

		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

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

		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

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

		// dump request body
		var reqBody []byte
		if config.WithRequestBody {
			buf, err := io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
				if len(buf) > RequestBodyMaxSize {
					reqBody = buf[:RequestBodyMaxSize]
				} else {
					reqBody = buf
				}
			}
		}

		// dump response body
		if config.WithResponseBody {
			c.Writer = newBodyWriter(c.Writer, ResponseBodyMaxSize)
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

		// request
		if config.WithRequestBody {
			attributes = append(attributes, slog.Group("request", slog.String("body", string(reqBody))))
		}
		if config.WithRequestHeader {
			for k, v := range c.Request.Header {
				if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
					continue
				}
				attributes = append(attributes, slog.Group("request", slog.Group("header", slog.Any(k, v))))
			}
		}

		// response
		if config.WithResponseBody {
			if w, ok := c.Writer.(*bodyWriter); ok {
				attributes = append(attributes, slog.Group("response", slog.String("body", w.body.String())))
			}
		}
		if config.WithResponseHeader {
			for k, v := range c.Writer.Header() {
				if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
					continue
				}
				attributes = append(attributes, slog.Group("response", slog.Group("header", slog.Any(k, v))))
			}
		}

		// custom context values
		if v, ok := c.Get(customAttributesCtxKey); ok {
			switch attrs := v.(type) {
			case []slog.Attr:
				attributes = append(attributes, attrs...)
			}
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

func AddCustomAttributes(c *gin.Context, attr slog.Attr) {
	v, exists := c.Get(customAttributesCtxKey)
	if !exists {
		c.Set(customAttributesCtxKey, []slog.Attr{attr})
		return
	}

	switch attrs := v.(type) {
	case []slog.Attr:
		c.Set(customAttributesCtxKey, append(attrs, attr))
	}
}
