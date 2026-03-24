package adapters

import (
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type SubscriptionModelAdapter struct {
	Subscription entities.EventSubscription
}

func (a SubscriptionModelAdapter) ToModel() models.Subscription {
	var projectID, instanceID uuid.UUID
	if a.Subscription.Project != nil {
		projectID = a.Subscription.Project.ID
	}
	if a.Subscription.Instance != nil {
		instanceID = a.Subscription.Instance.ID
	}
	return models.Subscription{
		Base: models.Base{
			ID:        a.Subscription.ID,
			CreatedAt: a.Subscription.CreatedAt,
		},
		ProjectID:  projectID,
		InstanceID: instanceID,
		NodeID: func() string {
			if a.Subscription.Node != nil {
				return a.Subscription.Node.ID
			}
			return ""
		}(),
		Type:           models.SubscriptionType(a.Subscription.Type),
		EventName:      a.Subscription.EventName,
		CorrelationKey: a.Subscription.CorrelationKey,
	}
}

type SubscriptionEntityAdapter struct {
	Model models.Subscription
}

func (a SubscriptionEntityAdapter) ToEntity() entities.EventSubscription {
	return entities.EventSubscription{
		ID:             a.Model.ID,
		Project:        &entities.Project{ID: a.Model.ProjectID},
		Instance:       &entities.ProcessInstance{ID: a.Model.InstanceID},
		Node:           &entities.Node{ID: a.Model.NodeID},
		Type:           entities.SubscriptionType(a.Model.Type),
		EventName:      a.Model.EventName,
		CorrelationKey: a.Model.CorrelationKey,
		CreatedAt:      a.Model.CreatedAt,
	}
}
