package model

import (
	"time"

	"github.com/gsoultan/storm"
)

// Token is a single point of execution inside an instance.
//
// Not a table: tokens live inside ProcessInstance.Tokens as jsonb, because they
// are read and rewritten as a set every time the instance advances — one row
// per token would be several writes where there is one.
type Token struct {
	ID         storm.UUID `json:"id"`
	InstanceID storm.UUID `json:"instance_id"`
	// NodeID is where this token sits. A token on a node the loaded definition
	// does not contain is the shape that used to hang the engine forever.
	NodeID string      `json:"node_id"`
	Status TokenStatus `json:"status"`
	// IterationID distinguishes the parallel copies of a multi-instance node.
	IterationID string         `json:"iteration_id,omitzero"`
	Variables   map[string]any `json:"variables,omitzero"`
	CreatedAt   time.Time      `json:"created_at,omitzero"`
}
