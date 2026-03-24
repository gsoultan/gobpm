package contracts

// ConditionEvaluator defines the interface for evaluating BPMN conditions.
type ConditionEvaluator interface {
	SetNext(next ConditionEvaluator) ConditionEvaluator
	Evaluate(condition string, vars map[string]any) bool
}
