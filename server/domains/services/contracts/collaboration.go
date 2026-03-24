package contracts

import (
	"context"
)

// CollaborationService handles real-time synchronization between users.
type CollaborationService interface {
	Broadcast(ctx context.Context, event any) error
}
