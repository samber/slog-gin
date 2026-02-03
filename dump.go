package sloggin

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var _ http.ResponseWriter = (*bodyWriter)(nil)
var _ http.Flusher = (*bodyWriter)(nil)
var _ http.Hijacker = (*bodyWriter)(nil)
var _ io.ReaderFrom = (*bodyWriter)(nil)

var errInvalidWrite = errors.New("invalid write result")

type bodyWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

// implements gin.ResponseWriter
func (w *bodyWriter) Write(b []byte) (int, error) {
	length := len(b)

	if w.body != nil {
		if w.body.Len()+length > w.maxSize {
			w.body.Truncate(min(w.maxSize, length, w.body.Len()))
			w.body.Write(b[:min(w.maxSize-w.body.Len(), length)])
		} else {
			w.body.Write(b)
		}
	}

	w.bytes += length //nolint:staticcheck
	return w.ResponseWriter.Write(b)
}

// implements io.ReaderFrom
func (w *bodyWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.body == nil {
		if rf, ok := w.ResponseWriter.(io.ReaderFrom); ok {
			n, err := rf.ReadFrom(r)
			w.bytes += int(n)
			return n, err
		}
	}
	// inline io.Copy(w, r) without ReaderFrom interface usage
	if wt, ok := r.(io.WriterTo); ok {
		return wt.WriteTo(w)
	}

	size := 32 * 1024
	if l, ok := r.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)

	var (
		written int64
		err     error
	)
	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := w.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}

func newBodyWriter(writer gin.ResponseWriter, maxSize int, recordBody bool) *bodyWriter {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}

	return &bodyWriter{
		ResponseWriter: writer,
		body:           body,
		maxSize:        maxSize,
		bytes:          0,
	}
}

type bodyReader struct {
	io.ReadCloser
	body    *bytes.Buffer
	maxSize int
	bytes   int
}

// implements io.Reader
func (r *bodyReader) Read(b []byte) (int, error) {
	n, err := r.ReadCloser.Read(b)
	if r.body != nil && r.body.Len() < r.maxSize {
		if r.body.Len()+n > r.maxSize {
			r.body.Write(b[:min(r.maxSize-r.body.Len(), n)])
		} else {
			r.body.Write(b[:n])
		}
	}
	r.bytes += n
	return n, err
}

func newBodyReader(reader io.ReadCloser, maxSize int, recordBody bool) *bodyReader {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}

	return &bodyReader{
		ReadCloser: reader,
		body:       body,
		maxSize:    maxSize,
		bytes:      0,
	}
}
