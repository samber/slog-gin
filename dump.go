package sloggin

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int
}

// implements gin.ResponseWriter
func (w bodyWriter) Write(b []byte) (int, error) {
	if w.body.Len()+len(b) > w.maxSize {
		w.body.Write(b[:w.maxSize-w.body.Len()])
	} else {
		w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func newBodyWriter(writer gin.ResponseWriter, maxSize int) *bodyWriter {
	return &bodyWriter{
		body:           bytes.NewBufferString(""),
		ResponseWriter: writer,
		maxSize:        maxSize,
	}
}
