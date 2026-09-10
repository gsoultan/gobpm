package repositories

import (
	"github.com/gsoultan/metis/server/repositories/contracts"
	stormdb "github.com/gsoultan/metis/server/repositories/db"
	"github.com/gsoultan/metis/server/repositories/gorms"
	"github.com/gsoultan/metis/server/repositories/pg"
	"gorm.io/gorm"
)

type gormRepository struct {
	audit                 contracts.AuditRepository
	broadcast             contracts.BroadcastRepository
	sharedCounter         contracts.SharedCounterRepository
	connector             contracts.ConnectorRepository
	connectorInstance     contracts.ConnectorInstanceRepository
	decision              contracts.DecisionRepository
	definition            contracts.DefinitionRepository
	environment           contracts.EnvironmentRepository
	deployment            contracts.DeploymentRepository
	externalTask          contracts.ExternalTaskRepository
	form                  contracts.FormRepository
	incident              contracts.IncidentRepository
	job                   contracts.JobRepository
	organization          contracts.OrganizationRepository
	process               contracts.ProcessRepository
	serviceCall           contracts.ServiceCallRepository
	webhook               contracts.WebhookRepository
	connectorManifest     contracts.ConnectorManifestRepository
	project               contracts.ProjectRepository
	subscription          contracts.SubscriptionRepository
	task                  contracts.TaskRepository
	user                  contracts.UserRepository
	group                 contracts.GroupRepository
	notification          contracts.NotificationRepository
	compensatableActivity contracts.CompensatableActivityRepository
	variableSnapshot      contracts.VariableSnapshotRepository
	uow                   contracts.UnitOfWork
}

// NewRepository creates a new composite repository.
//
// Two connections while the port is under way: the GORM one for the
// repositories that have not moved and the storm one for those that have. They
// are the same database — the composition root resolves the DSN once — which is
// what lets a repository move without the services calling it changing.
//
// A nil storm connection means the ported repositories are unavailable, which
// is a programming error rather than a configuration one now that PostgreSQL is
// the only engine. It panics rather than falling back to GORM: a fallback would
// mean the tests exercise one implementation and production the other.
func NewRepository(db *gorm.DB, conn *stormdb.Conn) Repository {
	if conn == nil {
		panic("repositories: a storm connection is required; the ported repositories have no GORM implementation left")
	}
	return &gormRepository{
		audit:                 pg.NewAuditRepository(conn),
		broadcast:             pg.NewBroadcastRepository(conn),
		sharedCounter:         pg.NewSharedCounterRepository(conn),
		connector:             pg.NewConnectorRepository(conn),
		connectorInstance:     pg.NewConnectorInstanceRepository(conn),
		decision:              gorms.NewDecisionRepository(db),
		definition:            gorms.NewDefinitionRepository(db),
		environment:           pg.NewEnvironmentRepository(conn),
		deployment:            pg.NewDeploymentRepository(conn),
		externalTask:          pg.NewExternalTaskRepository(conn),
		form:                  pg.NewFormRepository(conn),
		incident:              pg.NewIncidentRepository(conn),
		job:                   gorms.NewJobRepository(db),
		organization:          pg.NewOrganizationRepository(conn),
		process:               gorms.NewProcessRepository(db),
		serviceCall:           pg.NewServiceCallRepository(conn),
		webhook:               pg.NewWebhookRepository(conn),
		connectorManifest:     pg.NewConnectorManifestRepository(conn),
		project:               pg.NewProjectRepository(conn),
		subscription:          pg.NewSubscriptionRepository(conn),
		task:                  gorms.NewTaskRepository(db),
		user:                  gorms.NewUserRepository(db),
		group:                 gorms.NewGroupRepository(db),
		notification:          pg.NewNotificationRepository(conn),
		compensatableActivity: pg.NewCompensatableActivityRepository(conn),
		variableSnapshot:      pg.NewVariableSnapshotRepository(conn),
		uow:                   gorms.NewUnitOfWork(db),
	}
}

func (r *gormRepository) Audit() contracts.AuditRepository         { return r.audit }
func (r *gormRepository) Broadcast() contracts.BroadcastRepository { return r.broadcast }
func (r *gormRepository) SharedCounter() contracts.SharedCounterRepository {
	return r.sharedCounter
}
func (r *gormRepository) Connector() contracts.ConnectorRepository { return r.connector }
func (r *gormRepository) ConnectorInstance() contracts.ConnectorInstanceRepository {
	return r.connectorInstance
}
func (r *gormRepository) Decision() contracts.DecisionRepository         { return r.decision }
func (r *gormRepository) Definition() contracts.DefinitionRepository     { return r.definition }
func (r *gormRepository) Environment() contracts.EnvironmentRepository   { return r.environment }
func (r *gormRepository) Deployment() contracts.DeploymentRepository     { return r.deployment }
func (r *gormRepository) ExternalTask() contracts.ExternalTaskRepository { return r.externalTask }
func (r *gormRepository) Form() contracts.FormRepository                 { return r.form }
func (r *gormRepository) Incident() contracts.IncidentRepository         { return r.incident }
func (r *gormRepository) Job() contracts.JobRepository                   { return r.job }
func (r *gormRepository) Organization() contracts.OrganizationRepository { return r.organization }
func (r *gormRepository) Process() contracts.ProcessRepository           { return r.process }
func (r *gormRepository) ServiceCall() contracts.ServiceCallRepository   { return r.serviceCall }
func (r *gormRepository) Webhook() contracts.WebhookRepository           { return r.webhook }
func (r *gormRepository) ConnectorManifest() contracts.ConnectorManifestRepository {
	return r.connectorManifest
}
func (r *gormRepository) Project() contracts.ProjectRepository           { return r.project }
func (r *gormRepository) Subscription() contracts.SubscriptionRepository { return r.subscription }
func (r *gormRepository) Task() contracts.TaskRepository                 { return r.task }
func (r *gormRepository) User() contracts.UserRepository                 { return r.user }
func (r *gormRepository) Group() contracts.GroupRepository               { return r.group }
func (r *gormRepository) Notification() contracts.NotificationRepository { return r.notification }
func (r *gormRepository) CompensatableActivity() contracts.CompensatableActivityRepository {
	return r.compensatableActivity
}
func (r *gormRepository) VariableSnapshot() contracts.VariableSnapshotRepository {
	return r.variableSnapshot
}
func (r *gormRepository) UnitOfWork() contracts.UnitOfWork { return r.uow }
