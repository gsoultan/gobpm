package impl

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/internal/pkg/apierr"
	"github.com/gsoultan/metis/server/domains/entities"
	servicecontracts "github.com/gsoultan/metis/server/domains/services/contracts"
)

// NewUnavailableParticipantSyncService stands in when there is no storm
// connection, for the same reason the participant one does: a refusal that says
// why beats a nil that panics or an empty list that looks like no directories.
func NewUnavailableParticipantSyncService() servicecontracts.ParticipantSyncService {
	return unavailableSync{}
}

type unavailableSync struct{}

func (unavailableSync) ListSources(context.Context, uuid.UUID) ([]entities.ParticipantSource, error) {
	return nil, apierr.Invalidf("%s", unavailableReason)
}

func (unavailableSync) SaveSource(context.Context, entities.ParticipantSource) (uuid.UUID, error) {
	return uuid.Nil, apierr.Invalidf("%s", unavailableReason)
}

func (unavailableSync) DeleteSource(context.Context, uuid.UUID) error {
	return apierr.Invalidf("%s", unavailableReason)
}

func (unavailableSync) SyncSource(context.Context, uuid.UUID) (entities.ImportSummary, error) {
	return entities.ImportSummary{}, apierr.Invalidf("%s", unavailableReason)
}

// StartScheduledSyncs does nothing. There is nothing to schedule.
func (unavailableSync) StartScheduledSyncs(context.Context) {}
