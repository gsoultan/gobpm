package connector

import (
	"github.com/gsoultan/gobpm/server/domains/entities"
)

type ListConnectorsRequest struct{}
type ListConnectorsResponse struct {
	Connectors []entities.Connector `json:"connectors,omitzero"`
	Err        error                `json:"err,omitzero"`
}

func (r ListConnectorsResponse) Failed() error { return r.Err }

type CreateConnectorRequest struct {
	Connector entities.Connector `json:"connector"`
}
type CreateConnectorResponse struct {
	Connector entities.Connector `json:"connector,omitzero"`
	Err       error              `json:"err,omitzero"`
}

func (r CreateConnectorResponse) Failed() error { return r.Err }

type UpdateConnectorRequest struct {
	Connector entities.Connector `json:"connector"`
}
type UpdateConnectorResponse struct {
	Err error `json:"err,omitzero"`
}

func (r UpdateConnectorResponse) Failed() error { return r.Err }

type DeleteConnectorRequest struct {
	ID string `json:"id"`
}
type DeleteConnectorResponse struct {
	Err error `json:"err,omitzero"`
}

func (r DeleteConnectorResponse) Failed() error { return r.Err }

type ListConnectorInstancesRequest struct {
	ProjectID string `json:"project_id"`
}
type ListConnectorInstancesResponse struct {
	Instances []entities.ConnectorInstance `json:"instances,omitzero"`
	Err       error                        `json:"err,omitzero"`
}

func (r ListConnectorInstancesResponse) Failed() error { return r.Err }

type CreateConnectorInstanceRequest struct {
	Instance entities.ConnectorInstance `json:"instance"`
}
type CreateConnectorInstanceResponse struct {
	Instance entities.ConnectorInstance `json:"instance,omitzero"`
	Err      error                      `json:"err,omitzero"`
}

func (r CreateConnectorInstanceResponse) Failed() error { return r.Err }

type UpdateConnectorInstanceRequest struct {
	Instance entities.ConnectorInstance `json:"instance"`
}
type UpdateConnectorInstanceResponse struct {
	Err error `json:"err,omitzero"`
}

func (r UpdateConnectorInstanceResponse) Failed() error { return r.Err }

type DeleteConnectorInstanceRequest struct {
	ID string `json:"id"`
}
type DeleteConnectorInstanceResponse struct {
	Err error `json:"err,omitzero"`
}

func (r DeleteConnectorInstanceResponse) Failed() error { return r.Err }

type ExecuteConnectorRequest struct {
	ConnectorKey string         `json:"connector_key"`
	Config       map[string]any `json:"config"`
	Payload      map[string]any `json:"payload"`
}
type ExecuteConnectorResponse struct {
	Result map[string]any `json:"result,omitzero"`
	Err    error          `json:"err,omitzero"`
}

func (r ExecuteConnectorResponse) Failed() error { return r.Err }
