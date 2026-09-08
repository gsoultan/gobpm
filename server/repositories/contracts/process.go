package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/repositories/models"
)

// DefinitionInstanceCount is how much work one version of a process still
// holds. Running is what has to finish before that version has fully drained;
// Total is every instance it has ever had.
type DefinitionInstanceCount struct {
	Running int64
	Total   int64
}

// ProcessRepository defines the BPM process instance operations.
type ProcessRepository interface {
	Create(ctx context.Context, instance models.ProcessInstanceModel) (uuid.UUID, error)
	Get(ctx context.Context, id uuid.UUID) (models.ProcessInstanceModel, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (models.ProcessInstanceModel, error)
	Update(ctx context.Context, instance models.ProcessInstanceModel) error
	List(ctx context.Context) ([]models.ProcessInstanceModel, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.ProcessInstanceModel, error)

	// Paged variants for the instance list a user browses.
	ListPaged(ctx context.Context, p Pagination) (Page[models.ProcessInstanceModel], error)
	ListByProjectPaged(ctx context.Context, projectID uuid.UUID, p Pagination) (Page[models.ProcessInstanceModel], error)
	ListByDefinition(ctx context.Context, definitionID uuid.UUID) ([]models.ProcessInstanceModel, error)
	ListByParent(ctx context.Context, parentInstanceID uuid.UUID) ([]models.ProcessInstanceModel, error)
	CountByStatus(ctx context.Context, projectID uuid.UUID, status models.ProcessStatus) (int64, error)

	// CountInstancesByDefinitions reports how many instances each of the given
	// definition versions holds, in one grouped query rather than one per
	// version.
	//
	// It takes IDs rather than a process key because the caller has just listed
	// the versions, and because `key` is reserved on MySQL — keeping it out of
	// the predicate keeps this a plain IN over an indexed column.
	CountInstancesByDefinitions(ctx context.Context, definitionIDs []uuid.UUID) (map[uuid.UUID]DefinitionInstanceCount, error)
}
