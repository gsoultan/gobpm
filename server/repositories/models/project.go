package models

import (
	"github.com/google/uuid"
)

// ProjectModel represents the GORM model for projects.
type ProjectModel struct {
	Base
	OrganizationID uuid.UUID         `gorm:"type:uuid;index;uniqueIndex:idx_project_name_org" json:"organization_id,omitzero"`
	Organization   OrganizationModel `gorm:"foreignKey:OrganizationID" json:"organization,omitzero"`
	Name           string            `gorm:"uniqueIndex:idx_project_name_org" json:"name"`
	Description    string            `json:"description,omitzero"`
}

// TableName overrides the table name for ProjectModel.
func (ProjectModel) TableName() string {
	return "projects"
}
