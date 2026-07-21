package logging

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

type mockFailer struct {
	err error
}

type countingFailer struct {
	err   error
	calls int
}

func (f mockFailer) Failed() error {
	return f.err
}

func (f *countingFailer) Failed() error {
	f.calls++
	return f.err
}

func TestLoggingInterceptor(t *testing.T) {
	interceptor := NewLoggingInterceptor("test-method")

	t.Run("success", func(t *testing.T) {
		var e endpoint.Endpoint = func(ctx context.Context, request any) (any, error) {
			return "response", nil
		}
		e = interceptor.Intercept(e)
		_, err := e(t.Context(), "request")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("endpoint error", func(t *testing.T) {
		expectedErr := errors.New("endpoint error")
		var e endpoint.Endpoint = func(ctx context.Context, request any) (any, error) {
			return nil, expectedErr
		}
		e = interceptor.Intercept(e)
		_, err := e(t.Context(), "request")
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("failer error", func(t *testing.T) {
		expectedErr := errors.New("failer error")
		var e endpoint.Endpoint = func(ctx context.Context, request any) (any, error) {
			return mockFailer{err: expectedErr}, nil
		}
		e = interceptor.Intercept(e)
		_, err := e(t.Context(), "request")
		if err != nil {
			t.Errorf("expected no error from endpoint, got %v", err)
		}
	})

	t.Run("failer called once", func(t *testing.T) {
		expectedErr := errors.New("failer error")
		f := &countingFailer{err: expectedErr}
		var e endpoint.Endpoint = func(ctx context.Context, request any) (any, error) {
			return f, nil
		}

		e = interceptor.Intercept(e)
		_, err := e(t.Context(), "request")
		if err != nil {
			t.Errorf("expected no error from endpoint, got %v", err)
		}

		if f.calls != 1 {
			t.Errorf("expected Failed to be called once, got %d", f.calls)
		}
	})
}
