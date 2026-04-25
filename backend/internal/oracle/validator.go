package oracle

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	minCreativity = 0
	maxCreativity = 10
	maxImageBytes = 5 * 1024 * 1024 // 5MB
	maxNameLength = 64
)

var (
	allowedMIMEs = map[string]struct{}{
		"image/jpeg": {},
		"image/png":  {},
		"image/webp": {},
	}

	scriptTagPattern = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)
	urlPattern       = regexp.MustCompile(`(?i)https?://\S+`)
	disallowedChars  = regexp.MustCompile(`[^a-zA-Z0-9\s\-_'.,]`)
)

// ValidationError represents a structured validation failure for user inputs.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateRequest verifies the request payload and sanitizes the name field.
func ValidateRequest(req *OracleRequest) error {
	if req == nil {
		return errors.New("request cannot be nil")
	}

	req.Name = sanitizeName(req.Name)
	if req.Name == "" {
		return ValidationError{Field: "name", Message: "must contain letters, numbers, or basic punctuation"}
	}

	if req.Creativity < minCreativity || req.Creativity > maxCreativity {
		return ValidationError{Field: "creativity", Message: "must be between 0 and 10"}
	}

	if req.ImageBase64 == "" {
		return ValidationError{Field: "imageBase64", Message: "image data is required"}
	}

	if _, ok := allowedMIMEs[req.ImageMIME]; !ok {
		return ValidationError{Field: "imageMime", Message: "unsupported MIME type"}
	}

	data, err := base64.StdEncoding.DecodeString(req.ImageBase64)
	if err != nil {
		return ValidationError{Field: "imageBase64", Message: "must be valid base64"}
	}

	if len(data) == 0 {
		return ValidationError{Field: "imageBase64", Message: "image payload cannot be empty"}
	}

	if len(data) > maxImageBytes {
		return ValidationError{Field: "imageBase64", Message: "image payload exceeds 5MB limit"}
	}

	if req.ImageName == "" {
		return ValidationError{Field: "imageName", Message: "file name is required"}
	}

	return nil
}

func sanitizeName(input string) string {
	normalized := scriptTagPattern.ReplaceAllString(input, "")
	normalized = urlPattern.ReplaceAllString(normalized, "")
	normalized = disallowedChars.ReplaceAllString(normalized, "")
	normalized = strings.TrimSpace(normalized)
	if len(normalized) > maxNameLength {
		normalized = normalized[:maxNameLength]
	}
	return normalized
}
