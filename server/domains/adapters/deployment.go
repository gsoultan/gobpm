package adapters

import (
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type DeploymentModelAdapter struct {
	Deployment entities.Deployment
}

func (a DeploymentModelAdapter) ToModel() models.DeploymentModel {
	var projectID uuid.UUID
	if a.Deployment.Project != nil {
		projectID = a.Deployment.Project.ID
	}
	resources := make([]models.ResourceModel, len(a.Deployment.Resources))
	for i, r := range a.Deployment.Resources {
		var depID uuid.UUID
		if r.Deployment != nil {
			depID = r.Deployment.ID
		}
		resources[i] = models.ResourceModel{
			Base: models.Base{
				ID: r.ID,
			},
			DeploymentID: depID,
			Name:         r.Name,
			Content:      r.Content,
			Type:         r.Type,
		}
	}
	return models.DeploymentModel{
		Base: models.Base{
			ID:        a.Deployment.ID,
			CreatedAt: a.Deployment.CreatedAt,
		},
		ProjectID: projectID,
		Name:      a.Deployment.Name,
		Resources: resources,
	}
}

type DeploymentEntityAdapter struct {
	Model models.DeploymentModel
}

func (a DeploymentEntityAdapter) ToEntity() entities.Deployment {
	resources := make([]entities.Resource, len(a.Model.Resources))
	for i, r := range a.Model.Resources {
		resources[i] = entities.Resource{
			ID:         r.ID,
			Deployment: &entities.Deployment{ID: r.DeploymentID},
			Name:       r.Name,
			Content:    r.Content,
			Type:       r.Type,
		}
	}
	return entities.Deployment{
		ID:        a.Model.ID,
		Project:   &entities.Project{ID: a.Model.ProjectID},
		Name:      a.Model.Name,
		CreatedAt: a.Model.CreatedAt,
		Resources: resources,
	}
}
