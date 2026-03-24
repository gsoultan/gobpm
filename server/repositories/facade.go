package repositories

import "github.com/gsoultan/gobpm/server/repositories/contracts"

// Repository defines the composite repository interface.
type Repository interface {
	Audit() contracts.AuditRepository
	Connector() contracts.ConnectorRepository
	ConnectorInstance() contracts.ConnectorInstanceRepository
	Decision() contracts.DecisionRepository
	Definition() contracts.DefinitionRepository
	Deployment() contracts.DeploymentRepository
	ExternalTask() contracts.ExternalTaskRepository
	Form() contracts.FormRepository
	Incident() contracts.IncidentRepository
	Job() contracts.JobRepository
	Organization() contracts.OrganizationRepository
	Process() contracts.ProcessRepository
	Project() contracts.ProjectRepository
	Subscription() contracts.SubscriptionRepository
	Task() contracts.TaskRepository
	User() contracts.UserRepository
	Group() contracts.GroupRepository
	Notification() contracts.NotificationRepository
	CompensatableActivity() contracts.CompensatableActivityRepository
	VariableSnapshot() contracts.VariableSnapshotRepository
	UnitOfWork() contracts.UnitOfWork
}
