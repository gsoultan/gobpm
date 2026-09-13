package impl

import servicecontracts "github.com/gsoultan/metis/server/domains/services/contracts"

// WorkflowUserServiceForTest exposes the service to tests in other packages
// without widening the composition root's surface.
//
// It exists because the integration test builds the whole path itself —
// connection, repository, service — rather than going through the facade, which
// is what makes it a test of this repository rather than of the wiring.
type WorkflowUserServiceForTest struct {
	servicecontracts.WorkflowUserService
}
