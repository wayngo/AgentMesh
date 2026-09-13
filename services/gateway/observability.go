package main

import (
	"expvar"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log/slog"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type gatewayMetrics struct {
	requests      atomic.Uint64
	errors        atomic.Uint64
	latencyMicros atomic.Uint64
}

var metrics gatewayMetrics

func init() {
	expvar.Publish("agentmesh_gateway_requests_total", expvar.Func(func() any { return metrics.requests.Load() }))
	expvar.Publish("agentmesh_gateway_errors_total", expvar.Func(func() any { return metrics.errors.Load() }))
	expvar.Publish("agentmesh_gateway_latency_microseconds_total", expvar.Func(func() any { return metrics.latencyMicros.Load() }))
}

type observabilityResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *observabilityResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *observabilityResponseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
func ObservabilityMiddleware(next http.Handler) http.Handler {
	instrumented := otelhttp.NewHandler(next, "agentmesh.gateway")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &observabilityResponseWriter{ResponseWriter: w}
		metrics.requests.Add(1)
		instrumented.ServeHTTP(rw, r)
		elapsed := time.Since(start)
		metrics.latencyMicros.Add(uint64(elapsed.Microseconds()))
		if rw.status >= 500 {
			metrics.errors.Add(1)
		}
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "status", rw.status, "duration_ms", float64(elapsed.Microseconds())/1000, "request_id", r.Header.Get("X-Request-ID"), "status_code", strconv.Itoa(rw.status))
	})
}
