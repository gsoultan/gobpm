package grpcs

import (
	"github.com/gsoultan/gobpm/api/proto/services"
	"github.com/gsoultan/gobpm/server/endpoints"
	"github.com/gsoultan/gobpm/server/transports/grpcs/definitions"
	"github.com/gsoultan/gobpm/server/transports/grpcs/external_tasks"
	"github.com/gsoultan/gobpm/server/transports/grpcs/organizations"
	"github.com/gsoultan/gobpm/server/transports/grpcs/processes"
	"github.com/gsoultan/gobpm/server/transports/grpcs/projects"
	"github.com/gsoultan/gobpm/server/transports/grpcs/signals"
	"github.com/gsoultan/gobpm/server/transports/grpcs/stats"
	"github.com/gsoultan/gobpm/server/transports/grpcs/tasks"
)

type grpcServer struct {
	services.OrganizationServiceServer
	services.ProjectServiceServer
	services.ProcessServiceServer
	services.TaskServiceServer
	services.DefinitionServiceServer
	services.StatsServiceServer
	services.ExternalTaskServiceServer
	services.SignalServiceServer
}

func NewGRPCServer(eps endpoints.Endpoints) any {
	return &grpcServer{
		OrganizationServiceServer: organizations.NewServer(eps.Organization),
		ProjectServiceServer:      projects.NewServer(eps.Project),
		ProcessServiceServer:      processes.NewServer(eps.Process),
		TaskServiceServer:         tasks.NewServer(eps.Task),
		DefinitionServiceServer:   definitions.NewServer(eps.Definition),
		StatsServiceServer:        stats.NewServer(eps.Process),
		ExternalTaskServiceServer: external_tasks.NewServer(eps.ExternalTask),
		SignalServiceServer:       signals.NewServer(eps.Process),
	}
}
