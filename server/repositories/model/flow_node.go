package model

// FlowNode is one element of a process graph.
//
// Not a table: the whole graph lives inside ProcessDefinition.Nodes as jsonb.
// It nests — a sub-process carries its own nodes and flows — which is why it is
// stored as a document rather than as rows.
type FlowNode struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Type NodeType `json:"type"`

	// Who the work is offered to. Names and usernames, not foreign keys: a
	// process is authored against people who may not have accounts yet.
	Assignee        string   `json:"assignee,omitzero"`
	CandidateUsers  []string `json:"candidate_users,omitzero"`
	CandidateGroups []string `json:"candidate_groups,omitzero"`

	Priority int    `json:"priority,omitzero"`
	DueDate  string `json:"due_date,omitzero"`
	FormKey  string `json:"form_key,omitzero"`

	DefaultFlow   string `json:"default_flow,omitzero"`
	Script        string `json:"script,omitzero"`
	ScriptFormat  string `json:"script_format,omitzero"`
	ExternalTopic string `json:"external_topic,omitzero"`
	Documentation string `json:"documentation,omitzero"`

	AttachedToRef  string `json:"attachedToRef,omitzero"`
	ParentID       string `json:"parent_id,omitzero"`
	CancelActivity bool   `json:"cancel_activity,omitzero"`

	// Multi-instance: none, parallel or sequential, and what it iterates over.
	MultiInstanceType   string `json:"multi_instance_type,omitzero"`
	LoopCardinality     int    `json:"loop_cardinality,omitzero"`
	Collection          string `json:"collection,omitzero"`
	ElementVariable     string `json:"element_variable,omitzero"`
	CompletionCondition string `json:"completion_condition,omitzero"`

	ErrorCode         string `json:"error_code,omitzero"`
	IsAdHoc           bool   `json:"is_ad_hoc,omitzero"`
	IsEventSubProcess bool   `json:"is_event_sub_process,omitzero"`

	Incoming []string `json:"incoming,omitzero"`
	Outgoing []string `json:"outgoing,omitzero"`
	X        int      `json:"x,omitzero"`
	Y        int      `json:"y,omitzero"`

	Condition string `json:"condition,omitzero"`
	// Properties is the free-form bag the designer writes and handlers read.
	// Everything in it is authored content, so every read of it is untrusted
	// input and takes the comma-ok form.
	Properties map[string]any `json:"properties,omitzero"`

	// A sub-process carries its own graph.
	Nodes []FlowNode     `json:"nodes,omitzero"`
	Flows []SequenceFlow `json:"flows,omitzero"`
}
