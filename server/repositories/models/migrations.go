package models

// MigrationModels returns the ordered list of GORM models that must be
// auto-migrated on every supported database backend.
//
// This single source-of-truth is consumed by both the normal application
// startup path (app.go) and the first-time setup wizard (setup.go) so that
// the two never drift apart.
func MigrationModels() []any {
	return []any{
		new(OrganizationModel),
		new(ProcessInstanceModel),
		new(TaskModel),
		new(ProcessDefinitionModel),
		new(ProjectModel),
		new(AuditModel),
		new(JobModel),
		new(IncidentModel),
		new(UserModel),
		new(GroupModel),
		new(MembershipModel),
		new(ExternalTaskModel),
		new(Subscription),
		new(DecisionDefinitionModel),
		new(Connector),
		new(ConnectorInstance),
		// NotificationModel must be present in all environments.
		// It was previously omitted from setup.go's migrateTargetDatabase,
		// causing runtime failures the first time a notification was written.
		new(NotificationModel),
	}
}
