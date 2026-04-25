package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCORSMiddlewareHandlesOptions checks that preflight requests end early with 204.
func TestCORSMiddlewareHandlesOptions(t *testing.T) {
	called := false
	h := CORSMiddleware("http://example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/oracle", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if called {
		t.Fatal("next handler should not be called on OPTIONS")
	}
}

// TestTimeoutMiddlewareCancelsContext verifies request context gets cancelled after timeout.
func TestTimeoutMiddlewareCancelsContext(t *testing.T) {
	done := make(chan struct{})
	h := TimeoutMiddleware(20 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(done)
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	go h.ServeHTTP(rec, req)

	select {
	case <-done:
		return
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context was not cancelled by timeout middleware")
	}
}
