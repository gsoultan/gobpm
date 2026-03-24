package entities

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTaskAssignment NotificationType = "TaskAssignment"
	NotificationTaskOverdue    NotificationType = "TaskOverdue"
	NotificationIncident       NotificationType = "Incident"
	NotificationSystem         NotificationType = "System"
)

type Notification struct {
	ID        uuid.UUID        `json:"id"`
	User      *User            `json:"user,omitzero"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"is_read"`
	Link      string           `json:"link,omitzero"`
	CreatedAt time.Time        `json:"created_at"`
	Project   *Project         `json:"project,omitzero"`
	Instance  *ProcessInstance `json:"instance,omitzero"`
}
