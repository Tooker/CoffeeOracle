package oracle

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/logger"
)

const (
	PromptVersion = "v5"

	defaultModel = "gpt-5.5-nano"
	responsesURL = "https://api.openai.com/v1/responses"
)

// Service manages interaction with OpenAI Responses API for generating fortunes.
type Service struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
}

// NewService constructs a Service.
func NewService(apiKey string) *Service {
	return &Service{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    responsesURL,
	}
}

// NewServiceWithHTTPClient allows injecting custom HTTP clients/base URLs for testing.
func NewServiceWithHTTPClient(apiKey string, client *http.Client, baseURL string) *Service {
	return &Service{
		httpClient: client,
		apiKey:     apiKey,
		baseURL:    baseURL,
	}
}

// StreamEvent represents a single SSE event emitted by the Responses API.
type StreamEvent struct {
	Type string
	Data string
}

// StreamFortune validates the request, calls the Responses API, and streams output chunks.
func (s *Service) StreamFortune(ctx context.Context, req *OracleRequest, consume func(StreamEvent) error) error {
	if s.apiKey == "" {
		return errors.New("OPENAI_API_KEY is not configured")
	}
	if err := ValidateRequest(req); err != nil {
		return err
	}

	body, err := buildResponsesPayload(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	logger.Info("OpenAI request started model=%s prompt_version=%s", defaultModel, PromptVersion)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("OpenAI request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		payload, _ := io.ReadAll(resp.Body)
		logger.Error("OpenAI error response (%d): %s", resp.StatusCode, strings.TrimSpace(string(payload)))
		if len(payload) > 0 {
			return fmt.Errorf("openai error %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
		}
		return fmt.Errorf("openai error %d", resp.StatusCode)
	}

	decoder := newSSEDecoder(resp.Body)
	emittedChunks := 0
	for {
		evt, err := decoder.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("OpenAI stream reached EOF emitted_chunks=%d", emittedChunks)
				break
			}
			logger.Error("OpenAI stream decoder error: %v", err)
			return err
		}

		switch evt.Type {
		case "response.output_text.delta":
			var payload responseDelta
			if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
				continue
			}
			if payload.Delta != "" {
				emittedChunks++
				if err := consume(StreamEvent{Type: evt.Type, Data: payload.Delta}); err != nil {
					logger.Error("Stream consumer failed: %v", err)
					return err
				}
			}
		case "response.error":
			var payload responseError
			if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
				return fmt.Errorf("openai error: %s", evt.Data)
			}
			logger.Error("OpenAI streamed error: %s", payload.Error.Message)
			return fmt.Errorf("openai error: %s", payload.Error.Message)
		case "response.completed":
			logger.Info("OpenAI stream completed emitted_chunks=%d", emittedChunks)
			return nil
		default:
			if evt.Data == "[DONE]" {
				logger.Info("OpenAI stream done marker received emitted_chunks=%d", emittedChunks)
				return nil
			}
		}
	}

	return nil
}

// responseDelta is the JSON shape for a text chunk from OpenAI streaming events.
type responseDelta struct {
	Delta string `json:"delta"`
}

// responseError is the JSON shape for streamed OpenAI error events.
type responseError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// sseDecoder parses Server-Sent Events (SSE) line-by-line from the HTTP response stream.
type sseDecoder struct {
	reader *bufio.Reader
}

// newSSEDecoder creates a decoder around a generic stream reader.
func newSSEDecoder(r io.Reader) *sseDecoder {
	return &sseDecoder{reader: bufio.NewReader(r)}
}

// sseEvent is one parsed SSE message (event type + message data).
type sseEvent struct {
	Type string
	Data string
}

// Next reads until one full SSE event is available.
// It returns io.EOF when no more events are present.
func (d *sseDecoder) Next() (sseEvent, error) {
	var eventType strings.Builder
	var data strings.Builder
	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) && (eventType.Len() > 0 || data.Len() > 0) {
				break
			}
			return sseEvent{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if eventType.Len() == 0 && data.Len() == 0 {
				continue
			}
			break
		}
		if strings.HasPrefix(line, "event:") {
			eventType.Reset()
			eventType.WriteString(strings.TrimSpace(line[6:]))
		} else if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimSpace(line[5:]))
		}
	}
	if eventType.Len() == 0 && data.Len() == 0 {
		return sseEvent{}, io.EOF
	}
	return sseEvent{Type: eventType.String(), Data: data.String()}, nil
}

// responsesContent models one content item inside the OpenAI Responses API payload.
type responsesContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// responsesMessage models one chat-like message in the Responses API input array.
type responsesMessage struct {
	Role    string             `json:"role"`
	Content []responsesContent `json:"content"`
}

// responsesReasoning configures optional reasoning effort for the model.
type responsesReasoning struct {
	Effort string `json:"effort,omitempty"`
}

// responsesRequest is the full request body sent to OpenAI /v1/responses.
type responsesRequest struct {
	Model     string              `json:"model"`
	Reasoning *responsesReasoning `json:"reasoning,omitempty"`
	Stream    bool                `json:"stream"`
	Input     []responsesMessage  `json:"input"`
}

// buildResponsesPayload transforms user input into the OpenAI API JSON payload.
// It embeds the uploaded image as data URL and constructs the German oracle prompt.
func buildResponsesPayload(req *OracleRequest) ([]byte, error) {
	imageDataURI := fmt.Sprintf("data:%s;base64,%s", req.ImageMIME, req.ImageBase64)
	modeInstructions := `- Modus: freie Kaffeeschaum-Lesung.
- Beginne mit einer knappen Deutung des sichtbaren Musters.
- Erstelle danach eine passende Orakel-Lesung aus dem Milchschaum.
- Schließe mit einem kurzen, handlungsnahen Rat.`
	if req.QuestionMode {
		modeInstructions = `- Modus: konkrete Frage beantworten.
- Beginne mit einer knappen Deutung des sichtbaren Musters.
- Gib danach eine direkte Antwort auf die gestellte Frage.
- Schließe mit einem kurzen Rat oder nächsten Schritt.
- Wenn die Frage aus dem Bild nicht direkt entscheidbar ist, formuliere eine deutende Antwort mit vorsichtigem Rat statt einer absoluten Vorhersage.`
	}

	developerPrompt := fmt.Sprintf(`# Rolle und Ziel
Du bist das weltbekannte Kaffeemilchschaum-Orakel mit mehreren hundert Jahren Erfahrung im Lesen von Milchschaum.

# Anweisungen
- Der Nutzer %q möchte eine Lesung erhalten.
- Er hat einen Esoterik-Wert von %d auf einer Skala von 0 bis 10 gewählt.
- Dem Modell wird ein Bild von einer Tasse geliefert.
- Prüfe zuerst, ob auf dem Bild Milchschaum zu erkennen ist.
- Falls kein Milchschaum zu erkennen ist, antworte exakt: %q
- Falls Milchschaum zu erkennen ist:
%s
- Nenne 1-3 sichtbare Formen, Kontraste oder Muster aus dem Milchschaum.
- Erfinde keine Details, die im Bild nicht plausibel sichtbar sind.
- Erwähne **nicht**, welchen Wert der Nutzer gewählt hat.
- Erwähne **nicht** den Esoterik-Wert in der Antwort.
- Gib **nur** das eigentliche Orakel aus.

# Kontext
- Die Lesung soll sich wie eine echte Deutung aus dem Kaffeemilchschaum anfühlen.
- Der Nutzername %q ist als Kontext gegeben.
- Die Deutung soll sich auf das gelieferte Bild der Tasse mit Milchschaum beziehen.
- Nutze den Esoterik-Wert nur zur Tonalität: %s

# Ausgabeformat
- Der Output soll als nett formatierter Markdown-Text erscheinen.
- Schreibe 120-220 Wörter.
- Verwende maximal 2 kurze Markdown-Abschnitte.

# Verbosität
- Formuliere stimmungsvoll, direkt und passend zum Charakter eines Orakels.
- Bei medizinischen, rechtlichen, finanziellen oder sicherheitsrelevanten Fragen: antworte symbolisch und vorsichtig, gib keine verbindliche Fachentscheidung.`, req.Name, req.Creativity, "Ich kann hier keinen Milchschaum erkennen.", modeInstructions, req.Name, creativityTone(req.Creativity))

	userContent := []responsesContent{
		{Type: "input_text", Text: buildUserContext(req)},
		{Type: "input_image", ImageURL: imageDataURI},
	}

	body := responsesRequest{
		Model:     defaultModel,
		Reasoning: &responsesReasoning{Effort: "low"},
		Stream:    true,
		Input: []responsesMessage{
			{
				Role: "developer",
				Content: []responsesContent{
					{Type: "input_text", Text: developerPrompt},
				},
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
	}
	return json.Marshal(body)
}

// creativityTone maps the slider value to a hidden style guide without exposing the number.
func creativityTone(value int) string {
	switch {
	case value <= 3:
		return "bodenständig, humorvoll, wenig mystisch"
	case value <= 7:
		return "ausgewogen, symbolisch, warm"
	default:
		return "dramatisch, poetisch, stark orakelhaft"
	}
}

// buildUserContext keeps user-provided intent separate from the developer rules.
func buildUserContext(req *OracleRequest) string {
	var out strings.Builder
	out.WriteString("Bitte deute dieses Kaffeeschaumbild")
	if req.Name != "" {
		out.WriteString(" für ")
		out.WriteString(req.Name)
	}
	out.WriteString(".")
	if req.QuestionMode && req.Question != "" {
		out.WriteString("\nKonkrete Frage: ")
		out.WriteString(req.Question)
	}
	return out.String()
}
