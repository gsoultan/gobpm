package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// ParticipantSyncService manages the directories a project keeps its
// participants in step with, and runs them.
type ParticipantSyncService interface {
	ListSources(ctx context.Context, projectID uuid.UUID) ([]entities.ParticipantSource, error)
	SaveSource(ctx context.Context, source entities.ParticipantSource) (uuid.UUID, error)
	DeleteSource(ctx context.Context, id uuid.UUID) error

	// SyncSource reads one source now.
	//
	// Under a lock keyed on the source, so a scheduled run and somebody
	// pressing the button do not read the same directory twice and race each
	// other's writes.
	SyncSource(ctx context.Context, id uuid.UUID) (entities.ImportSummary, error)

	// StartScheduledSyncs runs due sources until the context is cancelled.
	StartScheduledSyncs(ctx context.Context)
}
