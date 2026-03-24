package contracts

import (
	"context"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type TaskFilter struct {
	ProjectID *uuid.UUID
	Status    []models.TaskStatus
	Assignee  *string
	Priority  *int
}

// TaskRepository defines the BPM task operations.
type TaskRepository interface {
	Get(ctx context.Context, id uuid.UUID) (models.TaskModel, error)
	List(ctx context.Context) ([]models.TaskModel, error)
	ListWithFilters(ctx context.Context, filter TaskFilter) ([]models.TaskModel, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]models.TaskModel, error)
	ListByAssignee(ctx context.Context, assignee string) ([]models.TaskModel, error)
	ListByCandidates(ctx context.Context, userID string, groups []string) ([]models.TaskModel, error)
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]models.TaskModel, error)
	Update(ctx context.Context, task models.TaskModel) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.TaskStatus) error
	Create(ctx context.Context, task models.TaskModel) error
	CountByStatus(ctx context.Context, projectID uuid.UUID, status models.TaskStatus) (int64, error)
}
