package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/oracle"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/logger"
)

const maxUploadBytes = 10 * 1024 * 1024

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
		"oracle request received name=%q creativity=%d image_mime=%q image_name=%q remote=%q prompt_version=%s",
		req.Name,
		req.Creativity,
		req.ImageMIME,
		req.ImageName,
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
	err = h.svc.StreamFortune(ctx, &req, func(evt oracle.StreamEvent) error {
		if evt.Data == "" {
			return nil
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

	fmt.Fprint(w, "event: complete\ndata: done\n\n")
	flusher.Flush()
	logger.Info("oracle request completed name=%q", req.Name)
}

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

func (h *OracleHandler) parseJSON(r *http.Request) (oracle.OracleRequest, error) {
	defer r.Body.Close()
	var req oracle.OracleRequest
	dec := json.NewDecoder(io.LimitReader(r.Body, maxUploadBytes))
	if err := dec.Decode(&req); err != nil {
		return oracle.OracleRequest{}, err
	}
	return req, nil
}

func (h *OracleHandler) parseForm(r *http.Request) (oracle.OracleRequest, error) {
	if err := r.ParseForm(); err != nil {
		return oracle.OracleRequest{}, err
	}
	creativity, _ := strconv.Atoi(r.FormValue("creativity"))
	return oracle.OracleRequest{
		Name:       r.FormValue("name"),
		Creativity: creativity,
	}, nil
}

func (h *OracleHandler) parseMultipart(r *http.Request) (oracle.OracleRequest, error) {
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		return oracle.OracleRequest{}, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return oracle.OracleRequest{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return oracle.OracleRequest{}, err
	}

	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "image/jpeg"
	}

	processed, processedMIME, err := resizeAndPersistImage(data, mime)
	if err != nil {
		return oracle.OracleRequest{}, err
	}

	creativity, err := strconv.Atoi(r.FormValue("creativity"))
	if err != nil {
		creativity = 5
	}

	return oracle.OracleRequest{
		Name:        r.FormValue("name"),
		Creativity:  creativity,
		ImageName:   header.Filename,
		ImageMIME:   processedMIME,
		ImageBase64: base64.StdEncoding.EncodeToString(processed),
	}, nil
}

func resizeAndPersistImage(data []byte, mime string) ([]byte, string, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
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
	encodeFormat := imaging.PNG
	if strings.Contains(mime, "jpeg") || strings.Contains(mime, "jpg") {
		encodeFormat = imaging.JPEG
	}

	if err := imaging.Encode(buf, img, encodeFormat); err != nil {
		return nil, "", err
	}

	tmpFile, err := os.CreateTemp("", "coffee-oracle-*.img")
	if err == nil {
		_, _ = tmpFile.Write(buf.Bytes())
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}

	outputMime := "image/png"
	if encodeFormat == imaging.JPEG {
		outputMime = "image/jpeg"
	}

	return buf.Bytes(), outputMime, nil
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func writeStreamError(w http.ResponseWriter, flusher http.Flusher, err error) {
	message := strings.ReplaceAll(err.Error(), "\n", " ")
	_ = writeSSEEvent(w, "response.error", message)
	flusher.Flush()
}

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
