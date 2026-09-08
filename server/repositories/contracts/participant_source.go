package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// ParticipantSourceRepository stores the directories a project syncs from.
type ParticipantSourceRepository interface {
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]entities.ParticipantSource, error)
	Get(ctx context.Context, id uuid.UUID) (entities.ParticipantSource, error)

	// GetWithSecrets returns a source with its credentials intact, for the
	// syncer. Separate from Get so that returning secrets is a deliberate call
	// rather than something every reader gets by default.
	GetWithSecrets(ctx context.Context, id uuid.UUID) (entities.ParticipantSource, error)

	Save(ctx context.Context, source entities.ParticipantSource) (uuid.UUID, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// DueForSync returns the enabled, scheduled sources whose next run has
	// arrived, across every project. Called by the worker, which has no request
	// and therefore no tenant.
	DueForSync(ctx context.Context) ([]entities.ParticipantSource, error)

	// RecordRun saves how a sync went.
	RecordRun(ctx context.Context, id uuid.UUID, run entities.SourceRun) error
}
