package contracts

import (
	"context"

	"github.com/google/uuid"
	"github.com/gsoultan/metis/server/domains/entities"
)

// MigrationService moves running instances from one version of a process onto
// another.
//
// The supported way to change version is to promote a new one and let the old
// one drain: an instance that finishes on the graph it started with cannot be
// broken by an edit. This is for the case drain cannot serve — work already in
// flight on a version that must not continue.
type MigrationService interface {
	// PlanInstanceMigration reports what MigrateInstances would do, and writes
	// nothing. Every refusal MigrateInstances would make appears here, so a
	// preview cannot approve something the apply rejects.
	PlanInstanceMigration(ctx context.Context, sourceDefID, targetDefID uuid.UUID, nodeMapping map[string]string) (entities.MigrationPlan, error)

	MigrateInstances(ctx context.Context, sourceDefID uuid.UUID, targetDefID uuid.UUID, nodeMapping map[string]string) error
}
