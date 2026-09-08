package httpclient

import (
	"errors"
	"strings"
	"testing"
)

func TestReadResponseBodyRefusesOverLimit(t *testing.T) {
	t.Setenv(envMaxResponseBytes, "16")
	_, err := ReadResponseBody(strings.NewReader(strings.Repeat("x", 64)))
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestReadResponseBodyAcceptsUpToLimit(t *testing.T) {
	t.Setenv(envMaxResponseBytes, "16")
	raw, err := ReadResponseBody(strings.NewReader(strings.Repeat("x", 16)))
	if err != nil {
		t.Fatalf("unexpected error at the limit: %v", err)
	}
	if len(raw) != 16 {
		t.Fatalf("read %d bytes, want 16", len(raw))
	}
}

func TestMaxResponseBytesDefaultAndOverride(t *testing.T) {
	t.Setenv(envMaxResponseBytes, "")
	if MaxResponseBytes() != defaultMaxResponseBytes {
		t.Fatalf("default not applied: %d", MaxResponseBytes())
	}
	t.Setenv(envMaxResponseBytes, "1024")
	if MaxResponseBytes() != 1024 {
		t.Fatalf("override not applied: %d", MaxResponseBytes())
	}
	t.Setenv(envMaxResponseBytes, "nonsense")
	if MaxResponseBytes() != defaultMaxResponseBytes {
		t.Fatalf("bad value should fall back to default: %d", MaxResponseBytes())
	}
}
