package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/logger"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/oracle"
)

const (
	maxUploadBytes  = 10 * 1024 * 1024
	imageStoreDir   = "data/oracle-images"
	readingStoreDir = "data/oracle-readings"
)

// OracleService abstracts the oracle domain logic for handler use.
type OracleService interface {
	StreamFortune(ctx context.Context, req *oracle.OracleRequest, consume func(oracle.StreamEvent) error) error
}

// OracleHandler handles POST /api/oracle requests.
type OracleHandler struct {
	svc OracleService
}

// NewOracleHandler constructs a handler.
func NewOracleHandler(svc OracleService) *OracleHandler {
	return &OracleHandler{svc: svc}
}

// ServeHTTP implements http.Handler.
func (h *OracleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	req, err := h.parseRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := oracle.ValidateRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	logger.Info(
		"oracle request received name=%q creativity=%d question_mode=%t image_mime=%q image_url=%q remote=%q prompt_version=%s",
		req.Name,
		req.Creativity,
		req.QuestionMode,
		req.ImageMIME,
		imageURL(req.ImageName),
		r.RemoteAddr,
		oracle.PromptVersion,
	)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, errors.New("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	var fortune strings.Builder
	err = h.svc.StreamFortune(ctx, &req, func(evt oracle.StreamEvent) error {
		if evt.Data == "" {
			return nil
		}
		if evt.Type == "response.output_text.delta" {
			fortune.WriteString(evt.Data)
		}
		if err := writeSSEEvent(w, evt.Type, evt.Data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})
	if err != nil {
		logger.Error("oracle request failed name=%q error=%v", req.Name, err)
		writeStreamError(w, flusher, err)
		return
	}

	shareURL, err := persistReading(fortune.String(), req.ImageName, req.Question)
	if err != nil {
		logger.Error("oracle share persistence failed name=%q error=%v", req.Name, err)
	} else if err := writeSSEJSONEvent(w, "share", map[string]string{"url": shareURL}); err != nil {
		logger.Error("oracle share event failed name=%q error=%v", req.Name, err)
		writeStreamError(w, flusher, err)
		return
	} else {
		logger.Info("oracle reading stored name=%q share_url=%q", req.Name, shareURL)
		flusher.Flush()
	}

	fmt.Fprint(w, "event: complete\ndata: done\n\n")
	flusher.Flush()
	logger.Info("oracle request completed name=%q", req.Name)
}

// parseRequest chooses the parser based on Content-Type, so clients can submit JSON, form, or multipart.
func (h *OracleHandler) parseRequest(r *http.Request) (oracle.OracleRequest, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return h.parseMultipart(r)
	}
	if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return h.parseForm(r)
	}
	return h.parseJSON(r)
}

// parseJSON handles application/json requests and guards body size.
func (h *OracleHandler) parseJSON(r *http.Request) (oracle.OracleRequest, error) {
	defer r.Body.Close()
	var req oracle.OracleRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, maxUploadBytes))
	if err := dec.Decode(&req); err != nil {
		return oracle.OracleRequest{}, err
	}
	return req, nil
}

// parseForm handles x-www-form-urlencoded requests (legacy/simple clients).
func (h *OracleHandler) parseForm(r *http.Request) (oracle.OracleRequest, error) {
	if err := r.ParseForm(); err != nil {
		return oracle.OracleRequest{}, err
	}
	creativity, _ := strconv.Atoi(r.FormValue("creativity"))
	return oracle.OracleRequest{
		Name:         r.FormValue("name"),
		Creativity:   creativity,
		QuestionMode: parseBool(r.FormValue("questionMode")),
		Question:     r.FormValue("question"),
	}, nil
}

// parseMultipart handles file uploads, processes image size/format, then builds the domain request.
func (h *OracleHandler) parseMultipart(r *http.Request) (oracle.OracleRequest, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return oracle.OracleRequest{}, err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return oracle.OracleRequest{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return oracle.OracleRequest{}, err
	}

	processed, processedMIME, imageName, err := resizeAndPersistImage(data)
	if err != nil {
		return oracle.OracleRequest{}, err
	}

	creativity, err := strconv.Atoi(r.FormValue("creativity"))
	if err != nil {
		creativity = 5
	}

	return oracle.OracleRequest{
		Name:         r.FormValue("name"),
		Creativity:   creativity,
		QuestionMode: parseBool(r.FormValue("questionMode")),
		Question:     r.FormValue("question"),
		ImageName:    imageName,
		ImageMIME:    processedMIME,
		ImageBase64:  base64.StdEncoding.EncodeToString(processed),
	}, nil
}

// parseBool accepts the HTML form values emitted by the frontend checkbox mode.
func parseBool(value string) bool {
	return value == "true" || value == "1" || value == "on"
}

// resizeAndPersistImage normalizes uploaded images to a reasonable max size.
// This protects API costs/performance and returns bytes + normalized MIME type.
func resizeAndPersistImage(data []byte) ([]byte, string, string, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", "", err
	}

	bounds := img.Bounds()
	maxSide := bounds.Dx()
	if bounds.Dy() > maxSide {
		maxSide = bounds.Dy()
	}

	if maxSide > 1024 {
		img = imaging.Fit(img, 1024, 1024, imaging.Lanczos)
	}

	buf := &bytes.Buffer{}
	if err := imaging.Encode(buf, img, imaging.JPEG); err != nil {
		return nil, "", "", err
	}

	imageName, err := persistImage(buf.Bytes())
	if err != nil {
		return nil, "", "", err
	}

	return buf.Bytes(), "image/jpeg", imageName, nil
}

// persistImage stores the processed image with a UUIDv4 filename for later log access.
func persistImage(data []byte) (string, error) {
	if err := os.MkdirAll(imageStoreDir, 0o755); err != nil {
		return "", err
	}

	id, err := uuid4String()
	if err != nil {
		return "", err
	}
	name := id + ".jpeg"

	if err := os.WriteFile(filepath.Join(imageStoreDir, name), data, 0o644); err != nil {
		return "", err
	}

	return name, nil
}

// persistReading stores the markdown fortune and returns its public share URL.
func persistReading(markdown string, imageName string, question string) (string, error) {
	if err := os.MkdirAll(readingStoreDir, 0o755); err != nil {
		return "", err
	}

	id, err := uuid4String()
	if err != nil {
		return "", err
	}

	content := buildReadingMarkdown(markdown, imageName, question)
	if err := os.WriteFile(filepath.Join(readingStoreDir, id+".md"), []byte(content), 0o644); err != nil {
		return "", err
	}

	return "/api/share/" + id, nil
}

// buildReadingMarkdown keeps the share artifact useful even outside the web UI.
func buildReadingMarkdown(markdown string, imageName string, question string) string {
	var out strings.Builder
	out.WriteString("# CoffeeOracle Lesung\n\n")
	if imageName != "" {
		out.WriteString("![Hochgeladenes Kaffeeschaumbild](")
		out.WriteString(imageURL(imageName))
		out.WriteString(")\n\n")
	}
	if strings.TrimSpace(question) != "" {
		out.WriteString("> Frage: ")
		out.WriteString(strings.TrimSpace(question))
		out.WriteString("\n\n")
	}
	out.WriteString(strings.TrimSpace(markdown))
	out.WriteString("\n")
	return out.String()
}

// uuid4String creates identifiers like "550e8400-e29b-41d4-a716-446655440000".
func uuid4String() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(raw)
	return fmt.Sprintf("%s-%s-%s-%s-%s", encoded[0:8], encoded[8:12], encoded[12:16], encoded[16:20], encoded[20:32]), nil
}

// imageURL returns the public API path for a stored image filename.
func imageURL(name string) string {
	if name == "" {
		return ""
	}
	return "/api/image/" + path.Base(name)
}

// writeError sends a standard JSON error response for non-streaming failures.
func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

// writeStreamError emits an SSE error event after the stream has already started.
func writeStreamError(w http.ResponseWriter, flusher http.Flusher, err error) {
	message := strings.ReplaceAll(err.Error(), "\n", " ")
	_ = writeSSEEvent(w, "response.error", message)
	flusher.Flush()
}

// writeSSEEvent formats one logical event in SSE wire format.
func writeSSEEvent(w http.ResponseWriter, eventType, data string) error {
	if eventType != "" {
		if _, err := fmt.Fprintf(w, "event: %s\n", eventType); err != nil {
			return err
		}
	}

	for _, line := range strings.Split(data, "\n") {
		if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "\n")
	return err
}

// writeSSEJSONEvent sends structured metadata over the existing SSE stream.
func writeSSEJSONEvent(w http.ResponseWriter, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return writeSSEEvent(w, eventType, string(data))
}
