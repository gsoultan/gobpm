package entities

import "slices"

// ProcessStatus defines the current state of a process instance.
type ProcessStatus string

const (
	ProcessActive    ProcessStatus = "active"
	ProcessCompleted ProcessStatus = "completed"
	ProcessSuspended ProcessStatus = "suspended"
	ProcessFailed    ProcessStatus = "failed"
)

var validProcessTransitions = map[ProcessStatus][]ProcessStatus{
	ProcessActive:    {ProcessCompleted, ProcessSuspended, ProcessFailed},
	ProcessSuspended: {ProcessActive, ProcessFailed},
	ProcessCompleted: {},
	ProcessFailed:    {},
}

func (s ProcessStatus) CanTransitionTo(target ProcessStatus) bool {
	return slices.Contains(validProcessTransitions[s], target)
}

// TaskStatus defines the current state of a task.
type TaskStatus string

const (
	TaskUnclaimed TaskStatus = "unclaimed" // same as pending
	TaskClaimed   TaskStatus = "claimed"
	TaskCompleted TaskStatus = "completed"
	TaskCanceled  TaskStatus = "canceled"
	TaskDelegated TaskStatus = "delegated"
	TaskEscalated TaskStatus = "escalated"
)

var validTaskTransitions = map[TaskStatus][]TaskStatus{
	TaskUnclaimed: {TaskClaimed, TaskCompleted, TaskCanceled, TaskEscalated},
	TaskClaimed:   {TaskUnclaimed, TaskCompleted, TaskCanceled, TaskDelegated, TaskEscalated},
	TaskDelegated: {TaskClaimed, TaskCompleted, TaskCanceled},
	TaskEscalated: {TaskClaimed, TaskCompleted, TaskCanceled},
	TaskCompleted: {},
	TaskCanceled:  {},
}

func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	return slices.Contains(validTaskTransitions[s], target)
}
