package entities

// NodeType represents the type of a BPMN element.
type NodeType string

const (
	StartEvent             NodeType = "startEvent"
	EndEvent               NodeType = "endEvent"
	UserTask               NodeType = "userTask"
	ServiceTask            NodeType = "serviceTask"
	ExclusiveGateway       NodeType = "exclusiveGateway"
	ParallelGateway        NodeType = "parallelGateway"
	InclusiveGateway       NodeType = "inclusiveGateway"
	ScriptTask             NodeType = "scriptTask"
	IntermediateCatchEvent NodeType = "intermediateCatchEvent"
	IntermediateThrowEvent NodeType = "intermediateThrowEvent"
	CallActivity           NodeType = "callActivity"
	ManualTask             NodeType = "manualTask"
	BusinessRuleTask       NodeType = "businessRuleTask"
	SubProcess             NodeType = "subProcess"
	BoundaryEvent          NodeType = "boundaryEvent"
	EventBasedGateway      NodeType = "eventBasedGateway"
	MessageEvent           NodeType = "messageEvent"
	SignalEvent            NodeType = "signalEvent"
	TimerEvent             NodeType = "timerEvent"
	ErrorEndEvent          NodeType = "errorEndEvent"
	TerminateEndEvent      NodeType = "terminateEndEvent"
	EscalationThrowEvent   NodeType = "escalationThrowEvent"
	CompensationThrowEvent NodeType = "compensationThrowEvent"
	Pool                   NodeType = "pool"
	Lane                   NodeType = "lane"
)
