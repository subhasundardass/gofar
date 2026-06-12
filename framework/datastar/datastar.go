package datastar

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/starfederation/datastar-go/datastar"
)

type fiberResponseWriter struct {
	c           *fiber.Ctx
	header      http.Header
	wroteHeader bool
	bw          *bufio.Writer
}

func (fw *fiberResponseWriter) Header() http.Header {
	if fw.header == nil {
		fw.header = make(http.Header)
	}
	return fw.header
}

func (fw *fiberResponseWriter) flushHeaders() {
	if fw.wroteHeader {
		return
	}
	fw.wroteHeader = true
	for k, vals := range fw.header {
		for _, v := range vals {
			fw.c.Response().Header.Set(k, v)
		}
	}
}

func (fw *fiberResponseWriter) Write(b []byte) (int, error) {
	fw.flushHeaders()
	n, err := fw.bw.Write(b)
	if err == nil {
		err = fw.bw.Flush()
	}
	return n, err
}

func (fw *fiberResponseWriter) WriteHeader(statusCode int) {
	fw.c.Status(statusCode)
	fw.flushHeaders()
}

func (fw *fiberResponseWriter) Flush() {
	if fw.bw != nil {
		_ = fw.bw.Flush()
	}
}
func fiberToHTTP(c *fiber.Ctx) (http.ResponseWriter, *http.Request, error) {
	r, err := http.NewRequest(
		c.Method(),
		c.OriginalURL(),
		bytes.NewReader(c.Body()),
	)
	if err != nil {
		return nil, nil, err
	}
	c.Request().Header.VisitAll(func(key, val []byte) {
		r.Header.Set(string(key), string(val))
	})
	return &fiberResponseWriter{c: c}, r, nil
}

func SSE(c *fiber.Ctx, fn func(*datastar.ServerSentEventGenerator) error) error {
	w, r, err := fiberToHTTP(c)
	if err != nil {
		return fmt.Errorf("fiberToHTTP: %w", err)
	}

	fw := w.(*fiberResponseWriter)
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	fw.wroteHeader = true

	var fnErr error
	c.Context().SetBodyStreamWriter(func(bw *bufio.Writer) {
		fw.bw = bw
		fnErr = fn(datastar.NewSSE(w, r))
	})

	return fnErr
}

// Unwrap returns the underlying writer so http.ResponseController can find
// the Flush method via interface unwrapping.
func (fw *fiberResponseWriter) Unwrap() http.ResponseWriter {
	return fw
}
