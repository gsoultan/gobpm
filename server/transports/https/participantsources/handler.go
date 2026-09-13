package participantsources

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gsoultan/metis/server/endpoints/participantsource"
	"github.com/gsoultan/metis/server/transports/https/common"
)

// RegisterHandlers mounts the participant directory registry.
func RegisterHandlers(m *http.ServeMux, eps participantsource.Endpoints, options []httptransport.ServerOption) {
	m.Handle("GET /api/v1/participant-sources", httptransport.NewServer(
		eps.ListSources,
		decodeListSourcesRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("POST /api/v1/participant-sources", httptransport.NewServer(
		eps.SaveSource,
		decodeSaveSourceRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("DELETE /api/v1/participant-sources/{id}", httptransport.NewServer(
		eps.DeleteSource,
		decodeDeleteSourceRequest,
		common.EncodeResponse,
		options...,
	))
	// Two segments with a wildcard first, which no other route here shares —
	// ServeMux refuses a pattern that could match the same path as another and
	// is no more specific, and it refuses it at registration, taking the whole
	// server down rather than one route.
	m.Handle("POST /api/v1/participant-sources/{id}/sync", httptransport.NewServer(
		eps.SyncSource,
		decodeSyncSourceRequest,
		common.EncodeResponse,
		options...,
	))
}

func decodeListSourcesRequest(_ context.Context, r *http.Request) (any, error) {
	return participantsource.ListSourcesRequest{ProjectID: r.URL.Query().Get("project_id")}, nil
}

func decodeSaveSourceRequest(_ context.Context, r *http.Request) (any, error) {
	var req participantsource.SaveSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeDeleteSourceRequest(_ context.Context, r *http.Request) (any, error) {
	return participantsource.DeleteSourceRequest{ID: r.PathValue("id")}, nil
}

func decodeSyncSourceRequest(_ context.Context, r *http.Request) (any, error) {
	return participantsource.SyncSourceRequest{ID: r.PathValue("id")}, nil
}
