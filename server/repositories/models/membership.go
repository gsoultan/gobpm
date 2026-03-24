package models

import "github.com/google/uuid"

// MembershipModel represents the GORM model for user-group membership.
type MembershipModel struct {
	UserID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id,omitzero"`
	GroupID uuid.UUID `gorm:"type:uuid;primaryKey" json:"group_id,omitzero"`
}

// TableName overrides the table name for MembershipModel.
func (MembershipModel) TableName() string {
	return "memberships"
}
