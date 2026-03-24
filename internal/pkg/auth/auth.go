package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

type AuthContextKey string

const (
	UserContextKey AuthContextKey = "user"
)

var (
	ErrUnauthorized         = errors.New("unauthorized: missing or invalid token")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

type UserClaims struct {
	Subject  string   `json:"sub"`
	Username string   `json:"preferred_username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

type TokenValidator struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

func NewTokenValidator(ctx context.Context, issuer string, clientID string) (*TokenValidator, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return &TokenValidator{
		provider: provider,
		verifier: verifier,
	}, nil
}

func (v *TokenValidator) ValidateToken(ctx context.Context, tokenString string) (*UserClaims, error) {
	idToken, err := v.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	var claims UserClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &claims, nil
}
