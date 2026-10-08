package sloggin

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type responseWrite struct {
	kind string
	body string
}

func writeResponse(t *testing.T, writer gin.ResponseWriter, writes []responseWrite) string {
	t.Helper()

	var body string
	for _, write := range writes {
		var n int
		var err error
		switch write.kind {
		case "bytes":
			n, err = writer.Write([]byte(write.body))
		case "string":
			n, err = writer.WriteString(write.body)
		case "io":
			n, err = io.WriteString(writer, write.body)
		default:
			t.Fatalf("unknown write kind %q", write.kind)
		}
		if err != nil {
			t.Fatalf("%s write returned an error: %v", write.kind, err)
		}
		if n != len(write.body) {
			t.Fatalf("%s write returned %d bytes, want %d", write.kind, n, len(write.body))
		}
		body += write.body
	}
	return body
}

func TestBodyWriterCapture(t *testing.T) {
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
		{"direct string without recording", []responseWrite{{"string", "pong"}}, 64, false},
		{"io string without recording", []responseWrite{{"io", "pong"}}, 64, false},
		{"zero cap", []responseWrite{{"string", "abc"}, {"bytes", "def"}}, 0, true},
		{"crossing byte cap", []responseWrite{{"bytes", "abcde"}, {"bytes", "XY"}}, 6, true},
		{"crossing mixed cap", []responseWrite{{"bytes", "abcde"}, {"string", "XY"}}, 6, true},
		{"full byte cap", []responseWrite{{"bytes", "abcdef"}, {"bytes", "XY"}}, 6, true},
		{"full mixed cap", []responseWrite{{"bytes", "abcdef"}, {"io", "XY"}}, 6, true},
		{"long initial string", []responseWrite{{"string", "abcdefgh"}, {"bytes", "X"}}, 6, true},
		{"UTF8 string", []responseWrite{{"string", "é界🙂"}}, 64, true},
		{"UTF8 byte cap", []responseWrite{{"io", "é界🙂"}}, 3, true},
		{"empty string after cap", []responseWrite{{"bytes", "abcdef"}, {"string", ""}}, 6, true},
		{"empty byte write after cap", []responseWrite{{"string", "abcdef"}, {"bytes", ""}}, 6, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			writer := newBodyWriter(ctx.Writer, test.maxSize, test.recordBody)

			body := writeResponse(t, writer, test.writes)
			if got := response.Body.String(); got != body {
				t.Errorf("response body = %q, want %q", got, body)
			}
			if writer.bytes != len(body) {
				t.Errorf("logged length = %d, want %d", writer.bytes, len(body))
			}
			if !test.recordBody {
				if writer.body != nil {
					t.Errorf("body buffer exists with recording disabled")
				}
				return
			}
			want := body[:min(len(body), test.maxSize)]
			if got := writer.body.String(); got != want {
				t.Errorf("captured body = %q, want %q", got, want)
			}
		})
	}
}
