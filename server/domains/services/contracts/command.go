package contracts

import "context"

// Command defines the interface for an executable process command.
type Command interface {
	Execute(ctx context.Context) error
}
