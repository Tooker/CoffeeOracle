package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestBuildResponsesPayloadIncludesQuestionMode verifies question mode rules and user context are separate.
func TestBuildResponsesPayloadIncludesQuestionMode(t *testing.T) {
	req := &OracleRequest{
		Name:         "Alex",
		Creativity:   5,
		QuestionMode: true,
		Question:     "Soll ich das neue Projekt wagen?",
		ImageName:    "cup.png",
		ImageMIME:    "image/png",
		ImageBase64:  "aGVsbG8=",
	}

	body, err := buildResponsesPayload(req)
	if err != nil {
		t.Fatalf("expected payload, got %v", err)
	}

	var payload responsesRequest
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("expected JSON payload, got %v", err)
	}

	developerPrompt := payload.Input[0].Content[0].Text
	userContext := payload.Input[1].Content[0].Text
	if strings.Contains(developerPrompt, req.Question) {
		t.Fatalf("expected developer prompt to avoid concrete question, got %s", developerPrompt)
	}
	if !strings.Contains(developerPrompt, "Gib danach eine direkte Antwort auf die gestellte Frage.") {
		t.Fatalf("expected question-mode instruction, got %s", developerPrompt)
	}
	if !strings.Contains(developerPrompt, "Nenne 1-3 sichtbare Formen") {
		t.Fatalf("expected image-grounding instruction, got %s", developerPrompt)
	}
	if !strings.Contains(developerPrompt, "Schreibe 120-220 Wörter") {
		t.Fatalf("expected length instruction, got %s", developerPrompt)
	}
	if !strings.Contains(userContext, req.Question) {
		t.Fatalf("expected user context to contain question, got %s", userContext)
	}
}

// TestServiceStreamFortuneSendsQuestionPayload checks the outbound request includes question prompt content.
func TestServiceStreamFortuneSendsQuestionPayload(t *testing.T) {
	requestBody := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		requestBody = string(raw)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "event: response.completed\ndata: {}\n\n")
	}))
	defer server.Close()

	svc := NewServiceWithHTTPClient("sk-test", server.Client(), server.URL)
	req := &OracleRequest{
		Name:         "Alex",
		Creativity:   5,
		QuestionMode: true,
		Question:     "Was soll ich beachten?",
		ImageName:    "cup.png",
		ImageMIME:    "image/png",
		ImageBase64:  "aGVsbG8=",
	}

	if err := svc.StreamFortune(context.Background(), req, func(evt StreamEvent) error { return nil }); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(requestBody, req.Question) {
		t.Fatalf("expected request body to include question, got %s", requestBody)
	}
}
