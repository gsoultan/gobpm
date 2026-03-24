package common

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gsoultan/gobpm/internal/pkg/auth"
	"github.com/gsoultan/gobpm/server/endpoints"
)

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response any) error {
	if f, ok := response.(endpoints.Failer); ok && f.Failed() != nil {
		EncodeError(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("EncodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(CodeFrom(err))
	json.NewEncoder(w).Encode(map[string]any{
		"error": err.Error(),
	})
}

func CodeFrom(err error) int {
	switch {
	case errors.Is(err, auth.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrAuthenticationFailed):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
