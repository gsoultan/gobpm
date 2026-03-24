package adapters

import (
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type ExternalTaskModelAdapter struct {
	ExternalTask entities.ExternalTask
}

func (a ExternalTaskModelAdapter) ToModel() models.ExternalTaskModel {
	var projectID, instanceID, defID uuid.UUID
	if a.ExternalTask.Project != nil {
		projectID = a.ExternalTask.Project.ID
	}
	if a.ExternalTask.ProcessInstance != nil {
		instanceID = a.ExternalTask.ProcessInstance.ID
	}
	if a.ExternalTask.ProcessDefinition != nil {
		defID = a.ExternalTask.ProcessDefinition.ID
	}
	return models.ExternalTaskModel{
		Base: models.Base{
			ID:        a.ExternalTask.ID,
			CreatedAt: a.ExternalTask.CreatedAt,
		},
		ProjectID:           projectID,
		ProcessInstanceID:   instanceID,
		ProcessDefinitionID: defID,
		NodeID: func() string {
			if a.ExternalTask.Node != nil {
				return a.ExternalTask.Node.ID
			}
			return ""
		}(),
		Topic:          a.ExternalTask.Topic,
		WorkerID:       a.ExternalTask.WorkerID,
		LockExpiration: a.ExternalTask.LockExpiration,
		Retries:        a.ExternalTask.Retries,
		RetryTimeout:   a.ExternalTask.RetryTimeout,
		ErrorMessage:   a.ExternalTask.ErrorMessage,
		ErrorDetails:   a.ExternalTask.ErrorDetails,
		Variables:      a.ExternalTask.Variables,
	}
}

type ExternalTaskEntityAdapter struct {
	Model models.ExternalTaskModel
}

func (a ExternalTaskEntityAdapter) ToEntity() entities.ExternalTask {
	return entities.ExternalTask{
		ID:                a.Model.ID,
		Project:           &entities.Project{ID: a.Model.ProjectID},
		ProcessInstance:   &entities.ProcessInstance{ID: a.Model.ProcessInstanceID},
		ProcessDefinition: &entities.ProcessDefinition{ID: a.Model.ProcessDefinitionID},
		Node:              &entities.Node{ID: a.Model.NodeID},
		Topic:             a.Model.Topic,
		WorkerID:          a.Model.WorkerID,
		LockExpiration:    a.Model.LockExpiration,
		Retries:           a.Model.Retries,
		RetryTimeout:      a.Model.RetryTimeout,
		ErrorMessage:      a.Model.ErrorMessage,
		ErrorDetails:      a.Model.ErrorDetails,
		Variables:         a.Model.Variables,
		CreatedAt:         a.Model.CreatedAt,
	}
}
