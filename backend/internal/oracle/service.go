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
	systemPrompt = "Du bist ein Orakel und liest professionell die Zukunft aus Milchschaum auf dem Kaffee. Antworte immer auf Deutsch. Antworte ausschliesslich als Markdown (kein Plain-Text-Format). Nutze immer diese Struktur: 1) '## Deutung', 2) '## Zeichen im Schaum' als Liste mit 3-5 Punkten, 3) '## Rat des Orakels' mit 2-3 konkreten Schritten. Nutze gezielte **Hervorhebungen**."
	defaultModel = "gpt-4o-mini"
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

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
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
	for {
		evt, err := decoder.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		switch evt.Type {
		case "response.output_text.delta":
			var payload responseDelta
			if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
				continue
			}
			if payload.Delta != "" {
				if err := consume(StreamEvent{Type: evt.Type, Data: payload.Delta}); err != nil {
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
			return nil
		default:
			if evt.Data == "[DONE]" {
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

type responsesRequest struct {
	Model  string             `json:"model"`
	Stream bool               `json:"stream"`
	Input  []responsesMessage `json:"input"`
}

func buildResponsesPayload(req *OracleRequest) ([]byte, error) {
	imageDataURI := fmt.Sprintf("data:%s;base64,%s", req.ImageMIME, req.ImageBase64)
	userPrompt := fmt.Sprintf("Was bedeutet diese Tasse fuer %s? Die gewuenschte Esoterik-Stufe ist %d von 10. Gib die Antwort exakt im angeforderten Markdown-Format mit den genannten Ueberschriften aus.", req.Name, req.Creativity)

	body := responsesRequest{
		Model:  defaultModel,
		Stream: true,
		Input: []responsesMessage{
			{
				Role: "developer",
				Content: []responsesContent{
					{Type: "input_text", Text: systemPrompt},
				},
			},
			{
				Role: "user",
				Content: []responsesContent{
					{Type: "input_image", ImageURL: imageDataURI},
					{Type: "input_text", Text: userPrompt},
				},
			},
		},
	}
	return json.Marshal(body)
}
