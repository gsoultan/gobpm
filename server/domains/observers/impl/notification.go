package impl

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/domains/observers/contracts"
	serviceContracts "github.com/gsoultan/gobpm/server/domains/services/contracts"
)

type notificationObserver struct {
	notificationService serviceContracts.NotificationService
}

func NewNotificationObserver(notificationService serviceContracts.NotificationService) contracts.ProcessObserver {
	return &notificationObserver{notificationService: notificationService}
}

func (o *notificationObserver) OnEvent(ctx context.Context, event entities.ProcessEvent) {
	switch event.Type {
	case entities.EventTaskCreated, entities.EventTaskClaimed:
		o.handleTaskEvent(ctx, event)
	}
}

func (o *notificationObserver) handleTaskEvent(ctx context.Context, event entities.ProcessEvent) {
	if event.Instance == nil || event.Project == nil {
		return
	}

	assignee, _ := event.Variables["assignee"].(string)
	if assignee == "" {
		return
	}

	taskName := "Task"
	if event.Node != nil && event.Node.Name != "" {
		taskName = event.Node.Name
	}
	var nodeID string
	if event.Node != nil {
		nodeID = event.Node.ID
	}

	notification := entities.Notification{
		ID:       uuid.New(),
		User:     &entities.User{Username: assignee},
		Type:     entities.NotificationTaskAssignment,
		Title:    "Task Update",
		Message:  fmt.Sprintf("Task '%s' in process '%s' needs your attention.", taskName, event.Instance.ID),
		Link:     fmt.Sprintf("/tasks?id=%s", nodeID),
		Project:  event.Project,
		Instance: event.Instance,
	}

	_ = o.notificationService.Send(ctx, notification)
}
