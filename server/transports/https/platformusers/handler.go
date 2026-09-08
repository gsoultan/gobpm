package platformusers

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gsoultan/metis/server/endpoints/platformuser"
	"github.com/gsoultan/metis/server/transports/https/common"
)

// RegisterHandlers mounts the platform account routes. Every one of them is
// administrative; the gate is applied where the endpoints are built.
func RegisterHandlers(m *http.ServeMux, eps platformuser.Endpoints, options []httptransport.ServerOption) {
	m.Handle("GET /api/v1/platform-users", httptransport.NewServer(
		eps.ListAccounts,
		decodeListAccountsRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("POST /api/v1/platform-users", httptransport.NewServer(
		eps.SaveAccount,
		decodeSaveAccountRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("DELETE /api/v1/platform-users/{id}", httptransport.NewServer(
		eps.DeleteAccount,
		decodeDeleteAccountRequest,
		common.EncodeResponse,
		options...,
	))
	m.Handle("PUT /api/v1/platform-users/{id}/roles", httptransport.NewServer(
		eps.SetRoles,
		decodeSetRolesRequest,
		common.EncodeResponse,
		options...,
	))
}

func decodeListAccountsRequest(context.Context, *http.Request) (any, error) {
	return nil, nil
}

func decodeSaveAccountRequest(_ context.Context, r *http.Request) (any, error) {
	var req platformuser.SaveAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	return req, nil
}

func decodeDeleteAccountRequest(_ context.Context, r *http.Request) (any, error) {
	return platformuser.DeleteAccountRequest{ID: r.PathValue("id")}, nil
}

// decodeSetRolesRequest takes the account from the path, not the body.
//
// The path is what the route matched and what an audit log records. A body that
// disagreed with it would be a request to change somebody else's roles wearing
// the URL of the account the caller was allowed to open.
func decodeSetRolesRequest(_ context.Context, r *http.Request) (any, error) {
	var req platformuser.SetRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}
	req.ID = r.PathValue("id")
	return req, nil
}
