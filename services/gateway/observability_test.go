package main

import (
	"expvar"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObservabilityMiddlewareRecordsRequest(t *testing.T) {
	before := metrics.requests.Load()
	handler := ObservabilityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 204 {
		t.Fatalf("status=%d", res.Code)
	}
	if metrics.requests.Load() != before+1 {
		t.Fatal("request metric not incremented")
	}
}
func TestMetricsEndpointIsAvailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/debug/metrics", nil)
	res := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.Handle("/debug/metrics", expvar.Handler())
	mux.ServeHTTP(res, req)
	if !strings.Contains(res.Body.String(), "agentmesh_gateway_requests_total") {
		t.Fatal("metric missing")
	}
}
