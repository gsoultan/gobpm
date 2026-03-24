package decision

import (
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
)

type ListDecisionsRequest struct {
	ProjectID string `json:"project_id,omitzero"`
}

type ListDecisionsResponse struct {
	Decisions []entities.DecisionDefinition `json:"decisions,omitzero"`
	Err       error                         `json:"err,omitzero"`
}

func (r ListDecisionsResponse) Failed() error { return r.Err }

type GetDecisionRequest struct {
	ID string `json:"id"`
}

type GetDecisionResponse struct {
	Decision entities.DecisionDefinition `json:"decision,omitzero"`
	Err      error                       `json:"err,omitzero"`
}

func (r GetDecisionResponse) Failed() error { return r.Err }

type CreateDecisionRequest struct {
	Decision entities.DecisionDefinition `json:"decision,omitzero"`
}

type CreateDecisionResponse struct {
	ID  uuid.UUID `json:"id"`
	Err error     `json:"err,omitzero"`
}

func (r CreateDecisionResponse) Failed() error { return r.Err }

type UpdateDecisionRequest struct {
	ID       string                      `json:"id"`
	Decision entities.DecisionDefinition `json:"decision,omitzero"`
}

type UpdateDecisionResponse struct {
	Err error `json:"err,omitzero"`
}

func (r UpdateDecisionResponse) Failed() error { return r.Err }

type DeleteDecisionRequest struct {
	ID string `json:"id"`
}

type DeleteDecisionResponse struct {
	Err error `json:"err,omitzero"`
}

func (r DeleteDecisionResponse) Failed() error { return r.Err }

type EvaluateDecisionRequest struct {
	Key       string         `json:"key"`
	Version   int            `json:"version,omitzero"`
	Variables map[string]any `json:"variables,omitzero"`
}

type EvaluateDecisionResponse struct {
	Result entities.DecisionResult `json:"result,omitzero"`
	Err    error                   `json:"err,omitzero"`
}

func (r EvaluateDecisionResponse) Failed() error { return r.Err }
