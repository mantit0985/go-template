package api

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mantit0985/go-template/src/core"
	"go.opentelemetry.io/otel/trace"
)

type contextKey struct{}

var (
	requestIDKey = contextKey{}
	realIPKey    = contextKey{}
)

type Middleware func(http.Handler) http.Handler

// Chain applies a sequence of middlewares to a handler.
func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

// RequestID adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RealIP extracts the real client IP address.
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Forwarded-For")
		if ipStr == "" {
			ipStr = r.Header.Get("X-Real-IP")
		}

		var clientIP netip.Addr
		if ipStr == "" {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil {
				clientIP, _ = netip.ParseAddr(host)
			}
		} else {
			ips := strings.Split(ipStr, ",")
			for i := len(ips) - 1; i >= 0; i-- {
				candidate := strings.TrimSpace(ips[i])
				parsedIP, err := netip.ParseAddr(candidate)
				if err != nil {
					continue
				}
				isTrusted := false
				for _, trusted := range core.AppConfig.TrustedProxies {
					if parsedIP == trusted {
						isTrusted = true
						break
					}
				}
				if !isTrusted {
					clientIP = parsedIP
					break
				}
			}
			if !clientIP.IsValid() {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					clientIP, _ = netip.ParseAddr(host)
				}
			}
		}
		ctx := context.WithValue(r.Context(), realIPKey, clientIP)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs the details of each request using slog.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		spanContext := trace.SpanContextFromContext(r.Context())
		var attrs []any
		if spanContext.HasTraceID() {
			attrs = append(attrs, slog.String("trace_id", spanContext.TraceID().String()))
		}

		rid, _ := r.Context().Value(requestIDKey).(string)
		if rid != "" {
			attrs = append(attrs, slog.String("request_id", rid))
		}

		logger := slog.Default().With(slog.Group("request", attrs...))

		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		logger.Info("request processed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.statusCode),
			slog.Duration("duration", time.Since(start)),
			slog.Int("size", ww.size),
		)
	})
}

// Recoverer recovers from panics and returns a 500 Internal Server Error.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Auth is a placeholder authentication middleware.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

// Compress handles Gzip compression.
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Del("Content-Length")
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			if err := gz.Close(); err != nil {
				slog.Error("failed to close gzip writer", slog.Any("error", err))
			}
			gzipPool.Put(gz)
		}()

		next.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not support hijacking")
}

type gzipResponseWriter struct {
	Writer *gzip.Writer
	http.ResponseWriter
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.Writer.Write(b)
}

func (grw gzipResponseWriter) WriteHeader(code int) {
	grw.ResponseWriter.WriteHeader(code)
}

func (grw gzipResponseWriter) Flush() {
	if err := grw.Writer.Flush(); err != nil {
		slog.Error("failed to flush gzip writer", slog.Any("error", err))
	}
	if f, ok := grw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (grw gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := grw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not support hijacking")
}
