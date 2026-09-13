package model

// ConnectorProperty is one field a connector needs configuring.
//
// Not a table: it lives inside Connector.Properties as jsonb. It is a plain
// struct so the repository can marshal it, and it is in its own file for the
// same reason every other struct here is.
type ConnectorProperty struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	// Type is one of string, password, boolean, number, select. A property whose
	// type is password is what the credential masking keys off.
	Type         string `json:"type"`
	Description  string `json:"description,omitzero"`
	DefaultValue string `json:"default_value,omitzero"`
	Required     bool   `json:"required,omitzero"`
	// Options are the choices, when Type is select.
	Options []any `json:"options,omitzero"`
}
