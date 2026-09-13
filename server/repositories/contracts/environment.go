package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/repositories/models"
)

// EnvironmentRepository stores the runtimes a project deploys into.
//
// These rows live in the main database. What each environment's own database
// holds is the runtime it owns — definitions, instances, tasks — and this
// repository never reaches into one; it only records how to reach it.
type EnvironmentRepository interface {
	Get(ctx context.Context, id uuid.UUID) (models.EnvironmentModel, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.EnvironmentModel, error)

	// ListAll returns every environment in the installation, unscoped.
	//
	// Called at boot to bind a listener per environment, before any request
	// exists to carry a tenant. Not reachable from a request.
	ListAll(ctx context.Context) ([]models.EnvironmentModel, error)

	Create(ctx context.Context, environment models.EnvironmentModel) error
	Update(ctx context.Context, environment models.EnvironmentModel) error
	Delete(ctx context.Context, id uuid.UUID) error

	// PortTaken reports whether another environment already claims a port.
	// Installation-wide, because two listeners cannot share one.
	PortTaken(ctx context.Context, port int, excluding uuid.UUID) (bool, error)
}
