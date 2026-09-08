package models

// Connector is the GORM model for Connector templates.
type Connector struct {
	Base
	Key         string              `gorm:"size:255;uniqueIndex" json:"key"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitzero"`
	Icon        string              `json:"icon,omitzero"`
	Type        string              `json:"type"`
	Schema      []ConnectorProperty `gorm:"type:text;serializer:json" json:"schema,omitzero"`
}

// ConnectorProperty defines the schema for a connector's configuration in the database.
type ConnectorProperty struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	Type         string `json:"type"` // string, password, boolean, number, select
	Description  string `json:"description,omitzero"`
	DefaultValue string `json:"default_value,omitzero"`
	Required     bool   `json:"required,omitzero"`
	Options      []any  `json:"options,omitzero"` // For select type
}

// ConnectorInstance is the GORM model for ConnectorInstance.
type ConnectorInstance struct {
	Base
	ProjectID   UUID   `gorm:"index" json:"project_id,omitzero"`
	ConnectorID UUID   `gorm:"index" json:"connector_id,omitzero"`
	Name        string `json:"name"`
	// Encrypted at rest. This map holds whatever a connector needs to
	// authenticate — a bearer token, an SMTP password, a signing secret — and it
	// was stored as plain JSON, so anybody with a database backup or a read
	// replica had every third-party credential in the installation. EncryptedMap
	// reads a row written before this change as cleartext and re-persists it
	// encrypted on the next write, so no migration is needed.
	Config EncryptedMap `gorm:"type:text" json:"config,omitzero"`
}
