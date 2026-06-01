package oracle

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

// TestValidateRequestSuccess verifies a valid payload passes all validation checks.
func TestValidateRequestSuccess(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0xFF}, 128))
	req := &OracleRequest{
		Name:        "Alex",
		Creativity:  7,
		ImageName:   "cup.png",
		ImageMIME:   "image/png",
		ImageBase64: payload,
	}

	if err := ValidateRequest(req); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if req.Name != "Alex" {
		t.Fatalf("expected sanitized name to remain 'Alex', got %s", req.Name)
	}
}

// TestValidateRequestAcceptsQuestionMode verifies concrete questions are sanitized and accepted.
func TestValidateRequestAcceptsQuestionMode(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := &OracleRequest{
		Name:         "Alex",
		Creativity:   7,
		QuestionMode: true,
		Question:     "<script>alert('x')</script> Soll ich das Projekt wagen? https://evil.test",
		ImageName:    "cup.png",
		ImageMIME:    "image/png",
		ImageBase64:  payload,
	}

	if err := ValidateRequest(req); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if req.Question != "Soll ich das Projekt wagen?" {
		t.Fatalf("unexpected sanitized question: %q", req.Question)
	}
}

// TestValidateRequestRequiresQuestionWhenEnabled blocks empty question mode submissions.
func TestValidateRequestRequiresQuestionWhenEnabled(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := &OracleRequest{
		Name:         "Alex",
		Creativity:   7,
		QuestionMode: true,
		ImageName:    "cup.png",
		ImageMIME:    "image/png",
		ImageBase64:  payload,
	}

	if err := ValidateRequest(req); err == nil {
		t.Fatal("expected error for missing question")
	}
}

// TestValidateRequestSanitizesName confirms unsafe name fragments are removed.
func TestValidateRequestSanitizesName(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := &OracleRequest{
		Name:        "<script>alert('x')</script>http://evil.com Robert!!",
		Creativity:  5,
		ImageName:   "cup.png",
		ImageMIME:   "image/png",
		ImageBase64: payload,
	}

	if err := ValidateRequest(req); err != nil {
		t.Fatalf("expected success after sanitization, got %v", err)
	}

	if strings.Contains(req.Name, "script") || strings.Contains(strings.ToLower(req.Name), "http") {
		t.Fatalf("expected sanitized name without script/url, got %s", req.Name)
	}
}

// TestValidateRequestCreativityRange ensures creativity outside 0..10 is rejected.
func TestValidateRequestCreativityRange(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := &OracleRequest{
		Name:        "Sam",
		Creativity:  11,
		ImageName:   "cup.png",
		ImageMIME:   "image/png",
		ImageBase64: payload,
	}

	if err := ValidateRequest(req); err == nil {
		t.Fatal("expected error for creativity > 10")
	}
}

// TestValidateRequestRejectsMime ensures unsupported image formats are blocked.
func TestValidateRequestRejectsMime(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte("hello"))
	req := &OracleRequest{
		Name:        "Sam",
		Creativity:  2,
		ImageName:   "cup.bmp",
		ImageMIME:   "image/bmp",
		ImageBase64: payload,
	}

	if err := ValidateRequest(req); err == nil {
		t.Fatal("expected error for unsupported MIME type")
	}
}

// TestValidateRequestRejectsLargePayload enforces upload size limits.
func TestValidateRequestRejectsLargePayload(t *testing.T) {
	large := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x01}, maxImageBytes+1))
	req := &OracleRequest{
		Name:        "Taylor",
		Creativity:  4,
		ImageName:   "big.png",
		ImageMIME:   "image/png",
		ImageBase64: large,
	}

	if err := ValidateRequest(req); err == nil {
		t.Fatal("expected error for oversized image payload")
	}
}

// TestValidateRequestRequiresImageData ensures empty image content is treated as invalid.
func TestValidateRequestRequiresImageData(t *testing.T) {
	req := &OracleRequest{
		Name:       "Taylor",
		Creativity: 4,
		ImageName:  "big.png",
		ImageMIME:  "image/png",
	}

	if err := ValidateRequest(req); err == nil {
		t.Fatal("expected error for missing base64 data")
	}
}
