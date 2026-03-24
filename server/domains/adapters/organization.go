package adapters

import (
	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type OrganizationModelAdapter struct {
	Organization entities.Organization
}

func (a OrganizationModelAdapter) ToModel() models.OrganizationModel {
	return models.OrganizationModel{
		Base: models.Base{
			ID:        a.Organization.ID,
			CreatedAt: a.Organization.CreatedAt,
			UpdatedAt: a.Organization.UpdatedAt,
		},
		Name:        a.Organization.Name,
		Description: a.Organization.Description,
	}
}

type OrganizationEntityAdapter struct {
	Model models.OrganizationModel
}

func (a OrganizationEntityAdapter) ToEntity() entities.Organization {
	return entities.Organization{
		ID:          a.Model.ID,
		Name:        a.Model.Name,
		Description: a.Model.Description,
		CreatedAt:   a.Model.CreatedAt,
		UpdatedAt:   a.Model.UpdatedAt,
	}
}
