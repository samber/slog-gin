package sloggin

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestResponseStringLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		writes     []responseWrite
		maxSize    int
		recordBody bool
	}{
		{"direct string", []responseWrite{{"string", "pong"}}, 64, true},
		{"io string", []responseWrite{{"io", "pong"}}, 64, true},
		{"mixed writes", []responseWrite{{"bytes", "ab"}, {"string", "cd"}, {"io", "é"}}, 64, true},
		{"recording disabled", []responseWrite{{"string", "pong"}}, 64, false},
		{"limited body", []responseWrite{{"bytes", "abcde"}, {"string", "XY"}, {"bytes", "Z"}}, 6, true},
		{"zero cap", []responseWrite{{"io", "pong"}}, 0, true},
		{"UTF8 string", []responseWrite{{"string", "é界🙂"}}, 64, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previousMaxSize := ResponseBodyMaxSize
			ResponseBodyMaxSize = test.maxSize
			t.Cleanup(func() { ResponseBodyMaxSize = previousMaxSize })

			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			config := DefaultConfig()
			config.WithRequestID = false
			config.WithResponseBody = test.recordBody
			router := gin.New()
			router.Use(NewWithConfig(logger, config))
			var body string
			router.GET("/text", func(ctx *gin.Context) {
				ctx.Status(http.StatusCreated)
				body = writeResponse(t, ctx.Writer, test.writes)
			})

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/text", nil))
			if response.Code != http.StatusCreated {
				t.Errorf("response status = %d, want %d", response.Code, http.StatusCreated)
			}
			if got := response.Body.String(); got != body {
				t.Errorf("response body = %q, want %q", got, body)
			}

			var entry struct {
				Response struct {
					Length int     `json:"length"`
					Body   *string `json:"body"`
					Status int     `json:"status"`
				} `json:"response"`
			}
			if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
				t.Fatalf("decode JSON log: %v", err)
			}
			if entry.Response.Length != len(body) {
				t.Errorf("logged length = %d, want %d", entry.Response.Length, len(body))
			}
			if entry.Response.Status != response.Code {
				t.Errorf("logged status = %d, want %d", entry.Response.Status, response.Code)
			}
			if !test.recordBody {
				if entry.Response.Body != nil {
					t.Errorf("response body was logged with recording disabled")
				}
				return
			}
			want := body[:min(len(body), test.maxSize)]
			if entry.Response.Body == nil {
				t.Fatal("response body is absent from the log")
			}
			if got := *entry.Response.Body; got != want {
				t.Errorf("logged body = %q, want %q", got, want)
			}
		})
	}
}
