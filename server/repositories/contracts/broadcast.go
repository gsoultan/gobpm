package contracts

import (
	"context"
	"time"

	"github.com/gsoultan/metis/server/domains/entities"
	"github.com/gsoultan/metis/server/repositories/models"
)

// BroadcastRepository is the shared bus that lets an SSE event produced on one
// replica reach browsers connected to another.
type BroadcastRepository interface {
	// Publish records one encoded event, with the scope it may be delivered in.
	//
	// The scope is stored rather than recomputed because the replica that
	// delivers the event has no context to recompute it from: the request that
	// produced it happened on another machine.
	Publish(ctx context.Context, origin string, scope entities.SSEScope, payload string) error

	// Since returns events newer than afterID that some *other* replica
	// produced, oldest first, at most limit of them. A replica delivers its own
	// events directly and so excludes them here.
	Since(ctx context.Context, origin string, afterID int64, limit int) ([]models.BroadcastEventModel, error)

	// LatestID is the sequence a replica starts from, so that a process
	// starting up delivers what happens next rather than replaying history to
	// browsers that were not there for it.
	LatestID(ctx context.Context) (int64, error)

	// Prune deletes events older than the cutoff. The table is a bus, not a
	// log: once every live replica has moved past a row it has no readers.
	Prune(ctx context.Context, olderThan time.Time) (int64, error)
}
