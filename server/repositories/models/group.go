package models

import (
	"github.com/google/uuid"
)

// GroupModel represents the GORM model for user groups.
type GroupModel struct {
	Base
	OrganizationID uuid.UUID `gorm:"type:uuid;index;uniqueIndex:idx_group_name_org" json:"organization_id,omitzero"`
	Name           string    `gorm:"uniqueIndex:idx_group_name_org" json:"name"`
	Description    string    `json:"description,omitzero"`
	Roles          []string  `gorm:"type:text;serializer:json" json:"roles,omitzero"`
}

// TableName overrides the table name for GroupModel.
func (GroupModel) TableName() string {
	return "groups"
}
