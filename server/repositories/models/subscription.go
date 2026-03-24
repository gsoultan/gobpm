package models

import (
	"github.com/google/uuid"
)

type SubscriptionType string

const (
	SubscriptionSignal  SubscriptionType = "signal"
	SubscriptionMessage SubscriptionType = "message"
)

type Subscription struct {
	Base
	ProjectID      uuid.UUID        `gorm:"type:uuid;index" json:"project_id,omitzero"`
	InstanceID     uuid.UUID        `gorm:"type:uuid;index" json:"instance_id,omitzero"`
	NodeID         string           `json:"node_id"`
	Type           SubscriptionType `gorm:"index" json:"type"`
	EventName      string           `gorm:"index" json:"event_name"`
	CorrelationKey string           `gorm:"index" json:"correlation_key,omitzero"`
}

func (Subscription) TableName() string {
	return "event_subscriptions"
}
