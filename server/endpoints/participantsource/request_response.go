package participantsource

import (
	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// ListSourcesRequest asks for a project's directories.
type ListSourcesRequest struct {
	ProjectID string `json:"project_id"`
}

type ListSourcesResponse struct {
	Sources []entities.ParticipantSource `json:"sources"`
	Err     error                        `json:"err,omitzero"`
}

func (r ListSourcesResponse) Failed() error { return r.Err }

// SaveSourceRequest creates a directory, or updates one when ID is set.
type SaveSourceRequest struct {
	ID        string `json:"id,omitzero"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	// Kind is http or postgres. A file upload is not a source: it is something
	// somebody did once, not somewhere to read again.
	Kind string `json:"kind"`
	// Config is the endpoint and headers, or the connection string and query.
	// A credential sent back as the masking sentinel means "keep what is
	// stored".
	Config map[string]any `json:"config,omitzero"`
	// Schedule is an ISO 8601 repeating interval — "R/PT1H". Empty means the
	// source runs only when somebody asks.
	Schedule string `json:"schedule,omitzero"`
	// OnMissing is leave or deactivate.
	OnMissing string `json:"on_missing,omitzero"`
	Enabled   bool   `json:"enabled"`
}

type SaveSourceResponse struct {
	ID  uuid.UUID `json:"id,omitzero"`
	Err error     `json:"err,omitzero"`
}

func (r SaveSourceResponse) Failed() error { return r.Err }

type DeleteSourceRequest struct {
	ID string `json:"id"`
}

type DeleteSourceResponse struct {
	Err error `json:"err,omitzero"`
}

func (r DeleteSourceResponse) Failed() error { return r.Err }

// SyncSourceRequest runs one directory now.
type SyncSourceRequest struct {
	ID string `json:"id"`
}

type SyncSourceResponse struct {
	Created     int                      `json:"created"`
	Updated     int                      `json:"updated"`
	Groups      int                      `json:"groups"`
	Deactivated int                      `json:"deactivated,omitzero"`
	Problems    []entities.ImportProblem `json:"problems,omitzero"`
	Err         error                    `json:"err,omitzero"`
}

func (r SyncSourceResponse) Failed() error { return r.Err }
