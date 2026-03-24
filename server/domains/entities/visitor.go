package entities

// DefinitionVisitor defines the interface for visiting BPMN elements.
type DefinitionVisitor interface {
	VisitDefinition(pd *ProcessDefinition)
	VisitFlowNode(n *Node)
	VisitSequenceFlow(sf *SequenceFlow)
}

// Acceptable defines the interface for elements that can accept a visitor.
type Acceptable interface {
	Accept(visitor DefinitionVisitor)
}
