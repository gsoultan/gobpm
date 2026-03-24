package entities

// Node represents a node in a BPMN process definition.
type Node struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Type                NodeType `json:"type"`
	Assignee            string   `json:"assignee,omitzero"`
	CandidateUsers      []*User  `json:"candidate_users,omitzero"`
	CandidateGroups     []*Group `json:"candidate_groups,omitzero"`
	Priority            int      `json:"priority,omitzero"`
	DueDate             string   `json:"due_date,omitzero"`
	FormKey             string   `json:"form_key,omitzero"`
	DefaultFlow         string   `json:"default_flow,omitzero"`
	Script              string   `json:"script,omitzero"`
	ScriptFormat        string   `json:"script_format,omitzero"`
	ExternalTopic       string   `json:"external_topic,omitzero"`
	Documentation       string   `json:"documentation,omitzero"`
	AttachedToRef       string   `json:"attached_to_ref,omitzero"`
	ParentID            string   `json:"parent_id,omitzero"`
	CancelActivity      bool     `json:"cancel_activity,omitzero"`
	MultiInstanceType   string   `json:"multi_instance_type,omitzero"` // none, parallel, sequential
	LoopCardinality     int      `json:"loop_cardinality,omitzero"`
	Collection          string   `json:"collection,omitzero"`
	ElementVariable     string   `json:"element_variable,omitzero"`
	CompletionCondition string   `json:"completion_condition,omitzero"`
	// ErrorCode is the BPMN errorCode on an error boundary event used to match CatchableError.
	// An empty ErrorCode catches all errors; a non-empty value catches only matching codes.
	ErrorCode         string          `json:"error_code,omitzero"`
	IsAdHoc           bool            `json:"is_ad_hoc,omitzero"`
	IsEventSubProcess bool            `json:"is_event_sub_process,omitzero"`
	Incoming          []string        `json:"incoming,omitzero"`
	Outgoing          []string        `json:"outgoing,omitzero"`
	X                 int             `json:"x,omitzero"`
	Y                 int             `json:"y,omitzero"`
	Condition         string          `json:"condition,omitzero"`
	Properties        map[string]any  `json:"properties,omitzero"`
	Nodes             []*Node         `json:"nodes,omitzero"`
	Flows             []*SequenceFlow `json:"flows,omitzero"`
}

func (n *Node) GetStringProperty(key string) string {
	if n.Properties == nil {
		return ""
	}
	if v, ok := n.Properties[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (n *Node) traverseFlows(callback func(*SequenceFlow)) {
	for i := range n.Flows {
		callback(n.Flows[i])
	}
	for i := range n.Nodes {
		n.Nodes[i].traverseFlows(callback)
	}
}

func (n *Node) Accept(visitor DefinitionVisitor) {
	visitor.VisitFlowNode(n)
	for i := range n.Nodes {
		n.Nodes[i].Accept(visitor)
	}
	for i := range n.Flows {
		visitor.VisitSequenceFlow(n.Flows[i])
	}
}

func (n *Node) traverse(callback func(*Node)) {
	callback(n)
	for i := range n.Nodes {
		n.Nodes[i].traverse(callback)
	}
}
