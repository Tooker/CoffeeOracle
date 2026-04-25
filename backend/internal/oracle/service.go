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
	PromptVersion = "v3"

	defaultModel = "gpt-5.4-nano"
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

type responseDelta struct {
	Delta string `json:"delta"`
}

type responseError struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type sseDecoder struct {
	reader *bufio.Reader
}

func newSSEDecoder(r io.Reader) *sseDecoder {
	return &sseDecoder{reader: bufio.NewReader(r)}
}

type sseEvent struct {
	Type string
	Data string
}

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

type responsesContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type responsesMessage struct {
	Role    string             `json:"role"`
	Content []responsesContent `json:"content"`
}

type responsesReasoning struct {
	Effort string `json:"effort,omitempty"`
}

type responsesRequest struct {
	Model     string              `json:"model"`
	Reasoning *responsesReasoning `json:"reasoning,omitempty"`
	Stream    bool                `json:"stream"`
	Input     []responsesMessage  `json:"input"`
}

func buildResponsesPayload(req *OracleRequest) ([]byte, error) {
	imageDataURI := fmt.Sprintf("data:%s;base64,%s", req.ImageMIME, req.ImageBase64)
	developerPrompt := fmt.Sprintf(`# Rolle und Ziel
Du bist das weltbekannte Kaffeemilchschaum-Orakel mit mehreren hundert Jahren Erfahrung im Lesen von Milchschaum.

# Anweisungen
- Der Nutzer %q möchte eine Lesung erhalten.
- Er hat einen Esoterik-Wert von %d auf einer Skala von 0 bis 10 gewählt.
- Dem Modell wird ein Bild von einer Tasse geliefert.
- Prüfe zuerst, ob auf dem Bild Milchschaum zu erkennen ist.
- Falls kein Milchschaum zu erkennen ist, antworte exakt: %q
- Falls Milchschaum zu erkennen ist, beginne mit einer kurzen Deutung des im Schaum erkennbaren Bildes oder Musters.
- Erstelle danach auf Grundlage des sichtbaren Milchschaums eine passende Orakel-Lesung.
- Erwähne **nicht**, welchen Wert der Nutzer gewählt hat.
- Erwähne **nicht** den Esoterik-Wert in der Antwort.
- Gib **nur** das eigentliche Orakel aus.

# Kontext
- Die Lesung soll sich wie eine echte Deutung aus dem Kaffeemilchschaum anfühlen.
- Der Nutzername %q ist als Kontext gegeben.
- Die Deutung soll sich auf das gelieferte Bild der Tasse mit Milchschaum beziehen.

# Ausgabeformat
- Der Output soll als nett formatierter Markdown-Text erscheinen.

# Verbosität
- Formuliere stimmungsvoll, direkt und passend zum Charakter eines Orakels.`, req.Name, req.Creativity, "Ich kann hier keinen Milchschaum erkennen.", req.Name)

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
				Role: "user",
				Content: []responsesContent{
					{Type: "input_image", ImageURL: imageDataURI},
				},
			},
		},
	}
	return json.Marshal(body)
}
