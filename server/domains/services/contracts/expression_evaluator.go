package contracts

import "context"

// ExpressionEvaluator is a Strategy interface for evaluating expressions against
// a set of variables. Implementations can support JavaScript, FEEL (DMN), Groovy,
// or any other expression language without changing the calling code.
// Used in sequence flow conditions, completion conditions, and decision rules.
type ExpressionEvaluator interface {
	// Evaluate evaluates the expression string against the given variables and
	// returns the result. The result type depends on the expression (bool, number, string).
	Evaluate(ctx context.Context, expression string, variables map[string]any) (any, error)

	// EvaluateBool is a convenience method that evaluates an expression expected
	// to return a boolean. Returns false and an error if the result is not boolean.
	EvaluateBool(ctx context.Context, expression string, variables map[string]any) (bool, error)
}
