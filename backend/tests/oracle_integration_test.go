package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/oracle"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/server/handlers"
)

type mockOracleService struct{}

func (mockOracleService) StreamFortune(ctx context.Context, req *oracle.OracleRequest, consume func(oracle.StreamEvent) error) error {
	return consume(oracle.StreamEvent{Type: "response.output_text.delta", Data: "Hello"})
}

func TestOracleHandlerIntegration(t *testing.T) {
	body := `{"name":"A","creativity":5,"imageName":"cup.png","imageMime":"image/png","imageBase64":"aGVsbG8="}`
	req := httptest.NewRequest(http.MethodPost, "/api/oracle", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h := handlers.NewOracleHandler(mockOracleService{})
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Hello") {
		t.Fatalf("expected stream chunk, got %s", rec.Body.String())
	}
}
