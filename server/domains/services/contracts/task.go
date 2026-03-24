package contracts

import (
	"context"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
)

// TaskService defines the task management operations.
type TaskService interface {
	GetTask(ctx context.Context, id uuid.UUID) (entities.Task, error)
	ListTasks(ctx context.Context, projectID uuid.UUID) ([]entities.Task, error)
	ListTasksByAssignee(ctx context.Context, assignee string) ([]entities.Task, error)
	ListTasksByCandidates(ctx context.Context, userID string, groups []string) ([]entities.Task, error)
	ClaimTask(ctx context.Context, id uuid.UUID, userID string) error
	UnclaimTask(ctx context.Context, id uuid.UUID) error
	DelegateTask(ctx context.Context, id uuid.UUID, userID string) error
	CompleteTask(ctx context.Context, id uuid.UUID, userID string, vars map[string]any) error
	CreateTaskForNode(ctx context.Context, instance entities.ProcessInstance, node entities.Node) error
	UpdateTask(ctx context.Context, task entities.Task) error
	AssignTask(ctx context.Context, id uuid.UUID, userID string) error
}
