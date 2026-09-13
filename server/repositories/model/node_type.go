package model

// NodeType is the kind of a BPMN element — a user task, a gateway, an end
// event. Carried as a string column rather than a PostgreSQL enum: BPMN's
// element set grows, and adding a value should not be a migration.
type NodeType string

// TokenStatus is where a single point of execution stands.
type TokenStatus string

// ProcessStatus is the state of a whole instance.
type ProcessStatus string

// TaskStatus is the state of a human task.
type TaskStatus string

// JobType and JobStatus describe queued work — a timer, a service call.
type (
	JobType   string
	JobStatus string
)

// IncidentStatus is whether a failure is still open.
type IncidentStatus string

// SubscriptionType distinguishes a message subscription from a signal one.
type SubscriptionType string
