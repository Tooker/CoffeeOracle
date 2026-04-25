package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/oracle"
)

type fakeOracleService struct {
	req    *oracle.OracleRequest
	chunks []oracle.StreamEvent
	err    error
}

func (f *fakeOracleService) StreamFortune(ctx context.Context, req *oracle.OracleRequest, consume func(oracle.StreamEvent) error) error {
	f.req = req
	for _, evt := range f.chunks {
		if err := consume(evt); err != nil {
			return err
		}
	}
	return f.err
}

func TestOracleHandlerJSONSuccess(t *testing.T) {
	img := base64.StdEncoding.EncodeToString([]byte("image"))
	body := map[string]any{
		"name":        "Alex",
		"creativity":  5,
		"imageName":   "cup.png",
		"imageMime":   "image/png",
		"imageBase64": img,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/oracle", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	svc := &fakeOracleService{chunks: []oracle.StreamEvent{
		{Type: "response.output_text.delta", Data: "Hello"},
		{Type: "response.output_text.delta", Data: " World"},
	}}
	h := NewOracleHandler(svc)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "data: Hello") {
		t.Fatalf("expected stream chunk, got %s", rec.Body.String())
	}
}

func TestOracleHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/oracle", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	svc := &fakeOracleService{}
	h := NewOracleHandler(svc)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestOracleHandlerServiceError(t *testing.T) {
	img := base64.StdEncoding.EncodeToString([]byte("image"))
	body := map[string]any{
		"name":        "Alex",
		"creativity":  5,
		"imageName":   "cup.png",
		"imageMime":   "image/png",
		"imageBase64": img,
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/oracle", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	svc := &fakeOracleService{err: context.DeadlineExceeded}
	h := NewOracleHandler(svc)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for streaming error path, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "event: response.error") {
		t.Fatalf("expected error event, got %s", rec.Body.String())
	}
}
