package common

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncodeErrorRedactsSensitiveData(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	err := errors.New("database failure password=super-secret")

	EncodeError(t.Context(), err, recorder)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status code: got %d want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.String()
	if strings.Contains(body, "super-secret") {
		t.Fatalf("response body leaked sensitive value: %s", body)
	}
	if !strings.Contains(body, "***REDACTED***") {
		t.Fatalf("response body should contain redacted marker: %s", body)
	}
}
