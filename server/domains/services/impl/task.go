package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/gsoultan/gobpm/server/domains/adapters"
	"github.com/gsoultan/gobpm/server/domains/entities"
	servicecontracts "github.com/gsoultan/gobpm/server/domains/services/contracts"
	"github.com/gsoultan/gobpm/server/repositories"
	"github.com/gsoultan/gobpm/server/repositories/models"

	"github.com/google/uuid"
)

type taskService struct {
	repo   repositories.Repository
	engine servicecontracts.ExecutionEngine
}

func NewTaskService(
	repo repositories.Repository,
	engine servicecontracts.ExecutionEngine,
) servicecontracts.TaskService {
	return &taskService{
		repo:   repo,
		engine: engine,
	}
}

func (s *taskService) GetTask(ctx context.Context, id uuid.UUID) (entities.Task, error) {
	m, err := s.repo.Task().Get(ctx, id)
	if err != nil {
		return entities.Task{}, err
	}
	return adapters.TaskEntityAdapter{Model: m}.ToEntity(), nil
}

func (s *taskService) ListTasks(ctx context.Context, projectID uuid.UUID) ([]entities.Task, error) {
	var ms []models.TaskModel
	var err error
	if projectID != uuid.Nil {
		ms, err = s.repo.Task().ListByProject(ctx, projectID)
	} else {
		ms, err = s.repo.Task().List(ctx)
	}
	if err != nil {
		return nil, err
	}
	res := make([]entities.Task, len(ms))
	for i, m := range ms {
		res[i] = adapters.TaskEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (s *taskService) ListTasksByAssignee(ctx context.Context, assignee string) ([]entities.Task, error) {
	ms, err := s.repo.Task().ListByAssignee(ctx, assignee)
	if err != nil {
		return nil, err
	}
	res := make([]entities.Task, len(ms))
	for i, m := range ms {
		res[i] = adapters.TaskEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (s *taskService) ListTasksByCandidates(ctx context.Context, userID string, groups []string) ([]entities.Task, error) {
	ms, err := s.repo.Task().ListByCandidates(ctx, userID, groups)
	if err != nil {
		return nil, err
	}
	res := make([]entities.Task, len(ms))
	for i, m := range ms {
		res[i] = adapters.TaskEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (s *taskService) ClaimTask(ctx context.Context, id uuid.UUID, userID string) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		m, err := s.repo.Task().Get(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}
		task := adapters.TaskEntityAdapter{Model: m}.ToEntity()
		if task.Status != entities.TaskUnclaimed {
			return fmt.Errorf("task %s is not unclaimed (current status: %s)", id, task.Status)
		}

		// Validation: check if user is a candidate
		isCandidate := len(task.CandidateUsers) == 0
		for _, u := range task.CandidateUsers {
			if u.Username == userID {
				isCandidate = true
				break
			}
		}
		if !isCandidate {
			return fmt.Errorf("user %s is not a candidate for task %s", userID, id)
		}

		task.Status = entities.TaskClaimed
		task.Assignee = &entities.User{Username: userID}
		if err := s.repo.Task().Update(txCtx, adapters.TaskModelAdapter{Task: task}.ToModel()); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskClaimed,
			Instance:  task.Instance,
			Project:   task.Project,
			Node:      task.Node,
			Timestamp: time.Now().Unix(),
			Variables: map[string]any{"assignee": userID},
		})

		return nil
	})
}

func (s *taskService) UnclaimTask(ctx context.Context, id uuid.UUID) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		m, err := s.repo.Task().Get(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}
		task := adapters.TaskEntityAdapter{Model: m}.ToEntity()
		if task.Status != entities.TaskClaimed {
			return fmt.Errorf("task %s is not claimed", id)
		}
		task.Status = entities.TaskUnclaimed
		task.Assignee = nil
		if err := s.repo.Task().Update(txCtx, adapters.TaskModelAdapter{Task: task}.ToModel()); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskUpdated,
			Instance:  task.Instance,
			Project:   task.Project,
			Node:      task.Node,
			Timestamp: time.Now().Unix(),
			Variables: task.Variables,
		})

		return nil
	})
}

func (s *taskService) DelegateTask(ctx context.Context, id uuid.UUID, userID string) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		m, err := s.repo.Task().Get(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get task: %w", err)
		}
		task := adapters.TaskEntityAdapter{Model: m}.ToEntity()
		task.Status = entities.TaskDelegated
		task.Assignee = &entities.User{Username: userID}
		if err := s.repo.Task().Update(txCtx, adapters.TaskModelAdapter{Task: task}.ToModel()); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskUpdated,
			Instance:  task.Instance,
			Project:   task.Project,
			Node:      task.Node,
			Timestamp: time.Now().Unix(),
			Variables: task.Variables,
		})

		return nil
	})
}

func (s *taskService) CompleteTask(ctx context.Context, id uuid.UUID, userID string, vars map[string]any) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		m, err := s.repo.Task().Get(txCtx, id)
		if err != nil {
			return fmt.Errorf("failed to get task %s: %w", id, err)
		}
		task := adapters.TaskEntityAdapter{Model: m}.ToEntity()

		if task.Status == entities.TaskCompleted {
			return fmt.Errorf("task %s is already completed", id)
		}

		if task.Assignee != nil && task.Assignee.Username != userID {
			return fmt.Errorf("task %s is assigned to %s, but user %s tried to complete it", id, task.Assignee.Username, userID)
		}

		if err := s.repo.Task().UpdateStatus(txCtx, id, models.TaskStatus(entities.TaskCompleted)); err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}

		instance, err := s.engine.GetInstanceForUpdate(txCtx, task.Instance.ID)
		if err != nil {
			return err
		}

		if _, err := s.repo.Definition().Get(txCtx, instance.Definition.ID); err != nil {
			return err
		}

		for k, v := range vars {
			instance.SetVariable(k, v)
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskCompleted,
			Instance:  &instance,
			Project:   instance.Project,
			Node:      task.Node,
			Timestamp: time.Now().Unix(),
			Variables: vars,
		})

		// Better way to get def:
		fullDef, _ := s.engine.GetProcessDefinition(txCtx, instance.Definition.ID)

		return s.engine.Proceed(txCtx, &instance, fullDef, func() string {
			if task.Node != nil {
				return task.Node.ID
			}
			return ""
		}())
	})
}

func (s *taskService) CreateTaskForNode(ctx context.Context, instance entities.ProcessInstance, node entities.Node) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		idObj, _ := uuid.NewV7()
		status := entities.TaskUnclaimed
		var assignee *entities.User
		if node.Assignee != "" {
			status = entities.TaskClaimed
			assignee = &entities.User{Username: node.Assignee}
		}

		candidateUsers := node.CandidateUsers
		candidateGroups := node.CandidateGroups

		var dueDate *time.Time
		if node.DueDate != "" {
			if t, err := time.Parse(time.RFC3339, node.DueDate); err == nil {
				dueDate = &t
			}
		}

		task := entities.Task{
			ID:              idObj,
			Project:         instance.Project,
			Instance:        &instance,
			Node:            &node,
			Name:            node.Name,
			Description:     node.Documentation,
			Type:            node.Type,
			Status:          status,
			Assignee:        assignee,
			CandidateUsers:  candidateUsers,
			CandidateGroups: candidateGroups,
			Priority:        node.Priority,
			DueDate:         dueDate,
			FormKey:         node.FormKey,
			FormDefinition:  node.GetStringProperty("form_definition"),
			Variables:       instance.Variables,
			CreatedAt:       time.Now(),
		}

		if err := s.repo.Task().Create(txCtx, adapters.TaskModelAdapter{Task: task}.ToModel()); err != nil {
			return err
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskCreated,
			Instance:  &instance,
			Project:   instance.Project,
			Node:      &node,
			Timestamp: time.Now().Unix(),
			Variables: instance.Variables,
		})

		return nil
	})
}

func (s *taskService) UpdateTask(ctx context.Context, task entities.Task) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		// Ensure task exists before updating
		m, err := s.repo.Task().Get(txCtx, task.ID)
		if err != nil {
			return err
		}
		existing := adapters.TaskEntityAdapter{Model: m}.ToEntity()
		// UpdateConnectorInstance specific allowed fields
		existing.Name = task.Name
		existing.Priority = task.Priority
		existing.DueDate = task.DueDate

		if err := s.repo.Task().Update(txCtx, adapters.TaskModelAdapter{Task: existing}.ToModel()); err != nil {
			return err
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskUpdated,
			Instance:  existing.Instance,
			Project:   existing.Project,
			Node:      existing.Node,
			Timestamp: time.Now().Unix(),
			Variables: existing.Variables,
		})

		return nil
	})
}

func (s *taskService) AssignTask(ctx context.Context, id uuid.UUID, userID string) error {
	return s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		m, err := s.repo.Task().Get(txCtx, id)
		if err != nil {
			return err
		}
		task := adapters.TaskEntityAdapter{Model: m}.ToEntity()
		task.Assignee = &entities.User{Username: userID}
		task.Status = entities.TaskClaimed
		if err := s.repo.Task().Update(txCtx, adapters.TaskModelAdapter{Task: task}.ToModel()); err != nil {
			return err
		}

		s.engine.DispatchEvent(txCtx, entities.ProcessEvent{
			Type:      entities.EventTaskClaimed,
			Instance:  task.Instance,
			Project:   task.Project,
			Node:      task.Node,
			Timestamp: time.Now().Unix(),
			Variables: map[string]any{"assignee": userID},
		})

		return nil
	})
}
