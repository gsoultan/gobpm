package models

import (
	"github.com/google/uuid"
)

type DeploymentModel struct {
	Base
	ProjectID uuid.UUID       `gorm:"type:uuid;index" json:"project_id,omitzero"`
	Name      string          `json:"name"`
	Resources []ResourceModel `gorm:"foreignKey:DeploymentID" json:"resources,omitzero"`
}

func (DeploymentModel) TableName() string {
	return "deployments"
}

type ResourceModel struct {
	Base
	DeploymentID uuid.UUID `gorm:"type:uuid;index" json:"deployment_id,omitzero"`
	Name         string    `json:"name"`
	Content      []byte    `json:"content,omitzero"`
	Type         string    `json:"type"`
}

func (ResourceModel) TableName() string {
	return "deployment_resources"
}
