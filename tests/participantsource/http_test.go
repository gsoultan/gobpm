package participantsource_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gsoultan/metis/server/domains/services/impl/participantsource"
)

func TestAnHTTPDirectoryIsFetchedAndValidated(t *testing.T) {
	// The egress guard blocks loopback by default, which is the whole point of
	// it — and it blocks a test server for the same reason it blocks a metadata
	// service. Opened here, for this test only, the way the connector tests do.
	t.Setenv("METIS_HTTP_ALLOW_PRIVATE_NETWORKS", "true")

	var gotAuth, gotAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
		  {"username":"ada","email":"ada@example.com","groups":["approvers"],"active":true},
		  {"username":"bob","email":"nope"}
		]}`))
	}))
	defer server.Close()

	source := participantsource.NewHTTPSource(participantsource.HTTPConfig{
		URL:     server.URL,
		Headers: map[string]string{"Authorization": "Bearer secret-token"},
	})
	result, err := source.Fetch(t.Context())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	if gotAuth != "Bearer secret-token" {
		t.Errorf("the configured headers should authenticate the call, got %q", gotAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("expected a JSON Accept header, got %q", gotAccept)
	}
	if len(result.Rows) != 1 || result.Rows[0].Username != "ada" {
		t.Fatalf("expected ada to import, got %+v", result.Rows)
	}
	// The same validation as a CSV, because it is the same code.
	if len(result.Problems) != 1 || result.Problems[0].Username != "bob" {
		t.Fatalf("bob's address should be refused, got %v", result.Problems)
	}
}

// The endpoint address is operator-supplied, so without a guard this is a
// request the server will make to any address somebody names — its own metadata
// service included. The guard is the shared one, and it runs on the resolved
// address rather than on the text of the URL.
func TestAnEndpointOnAPrivateNetworkIsRefused(t *testing.T) {
	t.Setenv("METIS_HTTP_ALLOW_PRIVATE_NETWORKS", "false")
	for _, tc := range []struct{ name, url string }{
		{"loopback", "http://127.0.0.1:80/users"},
		{"link-local metadata", "http://169.254.169.254/latest/meta-data/"},
		{"private range", "http://10.0.0.5/users"},
		{"not a url", "://nope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := participantsource.NewHTTPSource(participantsource.HTTPConfig{URL: tc.url})
			if _, err := source.Fetch(context.Background()); err == nil {
				t.Fatalf("%s should be refused", tc.name)
			}
		})
	}
}

// A failing endpoint is reported by status, not by quoting its body: somebody
// else's error page is not something to put in our logs and UI.
func TestAFailingEndpointIsReportedByStatus(t *testing.T) {
	t.Setenv("METIS_HTTP_ALLOW_PRIVATE_NETWORKS", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("<html><body>internal stack trace here</body></html>"))
	}))
	defer server.Close()

	source := participantsource.NewHTTPSource(participantsource.HTTPConfig{URL: server.URL})
	_, err := source.Fetch(t.Context())
	if err == nil {
		t.Fatal("a 500 should be reported as a failure")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("the failure should name the status, got %q", err)
	}
	if strings.Contains(err.Error(), "stack trace") {
		t.Errorf("the failure must not quote the body, got %q", err)
	}
}

// A response that is not a directory fails as a source, not as rows.
func TestANonDirectoryResponseFails(t *testing.T) {
	t.Setenv("METIS_HTTP_ALLOW_PRIVATE_NETWORKS", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer server.Close()

	source := participantsource.NewHTTPSource(participantsource.HTTPConfig{URL: server.URL})
	if _, err := source.Fetch(t.Context()); err == nil {
		t.Fatal("a response with no participant list should be refused")
	}
}
