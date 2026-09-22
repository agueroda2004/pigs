package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	platformhttp "server/internal/platform/http"
)

func TestCORS(t *testing.T) {
	allowedOrigin := "http://localhost:4200"
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	middleware := platformhttp.CORS([]string{allowedOrigin})

	t.Run("sets headers for an allowed origin", func(t *testing.T) {
		nextCalled = false
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", allowedOrigin)
		recorder := httptest.NewRecorder()

		middleware(next).ServeHTTP(recorder, request)

		if !nextCalled || recorder.Code != http.StatusOK {
			t.Fatalf("called=%v status=%d", nextCalled, recorder.Code)
		}
		if recorder.Header().Get("Access-Control-Allow-Origin") != allowedOrigin {
			t.Fatalf("allow origin = %q", recorder.Header().Get("Access-Control-Allow-Origin"))
		}
		if recorder.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("allow credentials = %q", recorder.Header().Get("Access-Control-Allow-Credentials"))
		}
		if recorder.Header().Get("Vary") != "Origin" {
			t.Fatalf("vary = %q", recorder.Header().Get("Vary"))
		}
	})

	t.Run("omits headers for a disallowed origin", func(t *testing.T) {
		nextCalled = false
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", "http://evil.example")
		recorder := httptest.NewRecorder()

		middleware(next).ServeHTTP(recorder, request)

		if !nextCalled || recorder.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("called=%v allow origin=%q", nextCalled, recorder.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("answers preflight without calling next", func(t *testing.T) {
		nextCalled = false
		request := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
		request.Header.Set("Origin", allowedOrigin)
		request.Header.Set("Access-Control-Request-Headers", "content-type")
		recorder := httptest.NewRecorder()

		middleware(next).ServeHTTP(recorder, request)

		if nextCalled || recorder.Code != http.StatusNoContent {
			t.Fatalf("called=%v status=%d", nextCalled, recorder.Code)
		}
		if recorder.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Fatal("expected allow methods header")
		}
		if recorder.Header().Get("Access-Control-Allow-Headers") != "content-type" {
			t.Fatalf("allow headers = %q", recorder.Header().Get("Access-Control-Allow-Headers"))
		}
	})

	t.Run("passes through requests without an origin", func(t *testing.T) {
		nextCalled = false
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		middleware(next).ServeHTTP(recorder, request)

		if !nextCalled || recorder.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("called=%v allow origin=%q", nextCalled, recorder.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}
