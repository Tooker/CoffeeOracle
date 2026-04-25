package oracle

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServiceStreamFortune checks end-to-end SSE decoding from mocked OpenAI stream.
func TestServiceStreamFortune(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: response.output_text.delta\ndata: {\"delta\":\"Hallo\"}\n\n")
		fmt.Fprintf(w, "event: response.completed\ndata: {}\n\n")
	}))
	defer server.Close()

	svc := NewServiceWithHTTPClient("sk-test", server.Client(), server.URL)
	req := &OracleRequest{
		Name:        "Alex",
		Creativity:  5,
		ImageName:   "cup.png",
		ImageMIME:   "image/png",
		ImageBase64: "aGVsbG8=",
	}

	var output string
	if err := svc.StreamFortune(context.Background(), req, func(evt StreamEvent) error {
		if evt.Type == "response.output_text.delta" {
			output += evt.Data
		}
		return nil
	}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output != "Hallo" {
		t.Fatalf("unexpected output: %s", output)
	}
}
