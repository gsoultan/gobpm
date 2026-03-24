package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system.
type User struct {
	ID            uuid.UUID       `json:"id"`
	Organizations []*Organization `json:"organizations,omitzero"`
	Projects      []*Project      `json:"projects,omitzero"`
	Username      string          `json:"username"`
	FullName      string          `json:"full_name"`
	DisplayName   string          `json:"display_name"`
	Organization  *Organization   `json:"organization,omitzero"`
	Email         string          `json:"email"`
	Roles         []string        `json:"roles"`
	CreatedAt     time.Time       `json:"created_at,omitzero"`
}
