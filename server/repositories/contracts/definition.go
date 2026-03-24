package contracts

import (
	"context"
	"github.com/gsoultan/gobpm/server/repositories/models"

	"github.com/google/uuid"
)

// DefinitionRepository defines the process definition operations.
type DefinitionRepository interface {
	Get(ctx context.Context, id uuid.UUID) (models.ProcessDefinitionModel, error)
	GetByKey(ctx context.Context, key string) (models.ProcessDefinitionModel, error)
	GetByKeyAndVersion(ctx context.Context, key string, version int) (models.ProcessDefinitionModel, error)
	List(ctx context.Context) ([]models.ProcessDefinitionModel, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProcessDefinitionModel, error)
	Create(ctx context.Context, definition models.ProcessDefinitionModel) error
	Delete(ctx context.Context, id uuid.UUID) error
}
