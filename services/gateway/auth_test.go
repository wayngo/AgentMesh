package main

import (
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewareAllowsHealthWithoutToken(t *testing.T) {
	h := NewAuthMiddleware("secret").Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	h := NewAuthMiddleware("secret").Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodGet, "/v1/runs", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user-1"})
	signed, _ := token.SignedString([]byte("secret"))
	h := NewAuthMiddleware("secret").Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subject, ok := SubjectFromContext(r.Context()); !ok || subject != "user-1" {
			t.Errorf("subject=%q ok=%v", subject, ok)
		}
		w.WriteHeader(204)
	}))
	r := httptest.NewRequest(http.MethodGet, "/v1/runs", nil)
	r.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestAuthMiddlewareRejectsWrongSignature(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "user-1"})
	signed, _ := token.SignedString([]byte("wrong"))
	h := NewAuthMiddleware("secret").Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodGet, "/v1/runs", nil)
	r.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
