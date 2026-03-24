package entities

import (
	"github.com/google/uuid"
	"time"
)

type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

type JobType string

const (
	JobServiceTask   JobType = "service_task"
	JobTimer         JobType = "timer"
	JobTimerBoundary JobType = "timer_boundary"
)

type Job struct {
	ID         uuid.UUID          `json:"id"`
	Instance   *ProcessInstance   `json:"instance,omitzero"`
	Definition *ProcessDefinition `json:"definition,omitzero"`
	Node       *Node              `json:"node,omitzero"`
	Type       JobType            `json:"type"`
	Status     JobStatus          `json:"status"`
	Payload    map[string]any     `json:"payload"`
	Retries    int                `json:"retries"`
	MaxRetries int                `json:"maxRetries"`
	NextRunAt  time.Time          `json:"next_run_at"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	LastError  string             `json:"last_error,omitzero"`
}
