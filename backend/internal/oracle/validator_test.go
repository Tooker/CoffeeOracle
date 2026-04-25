package oracle

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

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
