package impl

import (
	"context"

	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/domains/observers/impl"
	"github.com/gsoultan/metis/server/domains/services/contracts"
)

type collaborationService struct {
	sse *impl.SSEObserver
}

func NewCollaborationService(sse *impl.SSEObserver) contracts.CollaborationService {
	return &collaborationService{sse: sse}
}

// Broadcast sends a designer presence event to the people who may see it.
//
// The scope comes from the request: the organization from the token, the
// environment from the listener it arrived on. A collaboration event names a
// process model and who is editing it, which is not a fact another tenant is
// entitled to.
func (s *collaborationService) Broadcast(ctx context.Context, event any) error {
	s.sse.BroadcastTo(entities.SSEScopeFrom(ctx), event)
	return nil
}
