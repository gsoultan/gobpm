package impl

import (
	"context"
	"fmt"

	"github.com/gsoultan/gobpm/server/domains/entities"
	"github.com/gsoultan/gobpm/server/domains/services/contracts"
)

// DecisionTableEvaluatorImpl implements DecisionTableEvaluator by applying the
// table's hit policy against each rule using the injected ExpressionEvaluator Strategy.
// Supports UNIQUE, FIRST, COLLECT, ANY, and PRIORITY hit policies.
type DecisionTableEvaluatorImpl struct {
	expr contracts.ExpressionEvaluator
}

// NewDecisionTableEvaluator creates a new DecisionTableEvaluatorImpl.
func NewDecisionTableEvaluator(expr contracts.ExpressionEvaluator) contracts.DecisionTableEvaluator {
	return &DecisionTableEvaluatorImpl{expr: expr}
}

// EvaluateTable evaluates the decision table against the given variables.
func (e *DecisionTableEvaluatorImpl) EvaluateTable(ctx context.Context, def entities.DecisionDefinition, variables map[string]any) (entities.DecisionResult, error) {
	matched, err := e.collectMatchingRules(ctx, def, variables)
	if err != nil {
		return entities.DecisionResult{}, err
	}
	return e.applyHitPolicy(def, matched)
}

// collectMatchingRules returns all rules whose input conditions match the variables.
func (e *DecisionTableEvaluatorImpl) collectMatchingRules(ctx context.Context, def entities.DecisionDefinition, variables map[string]any) ([]entities.DecisionRule, error) {
	var matched []entities.DecisionRule
	for _, rule := range def.Rules {
		ok, err := e.ruleMatches(ctx, def, rule, variables)
		if err != nil {
			return nil, err
		}
		if ok {
			matched = append(matched, rule)
		}
	}
	return matched, nil
}

// ruleMatches returns true if all input conditions of the rule match the variables.
func (e *DecisionTableEvaluatorImpl) ruleMatches(ctx context.Context, def entities.DecisionDefinition, rule entities.DecisionRule, variables map[string]any) (bool, error) {
	for i, inputExpr := range rule.Inputs {
		if i >= len(def.Inputs) {
			break
		}
		inputDef := def.Inputs[i]
		inputVal := variables[inputDef.Expression]
		vars := map[string]any{"_input": inputVal}
		ok, err := e.expr.EvaluateBool(ctx, inputExpr, vars)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

// applyHitPolicy applies the table's hit policy to the matched rules.
func (e *DecisionTableEvaluatorImpl) applyHitPolicy(def entities.DecisionDefinition, matched []entities.DecisionRule) (entities.DecisionResult, error) {
	if len(matched) == 0 {
		return entities.DecisionResult{Values: map[string]any{}}, nil
	}
	switch def.HitPolicy {
	case entities.HitPolicyFirst:
		return e.buildResult(def, []entities.DecisionRule{matched[0]}), nil
	case entities.HitPolicyCollect:
		return e.applyCollect(def, matched), nil
	case entities.HitPolicyUnique, entities.HitPolicyAny, entities.HitPolicyPriority, "":
		if len(matched) > 1 && def.HitPolicy == entities.HitPolicyUnique {
			return entities.DecisionResult{}, fmt.Errorf("UNIQUE hit policy violated: %d rules matched", len(matched))
		}
		return e.buildResult(def, []entities.DecisionRule{matched[0]}), nil
	default:
		return e.buildResult(def, []entities.DecisionRule{matched[0]}), nil
	}
}

// applyCollect aggregates all matched rule outputs according to the aggregation function.
func (e *DecisionTableEvaluatorImpl) applyCollect(def entities.DecisionDefinition, matched []entities.DecisionRule) entities.DecisionResult {
	result := e.buildResult(def, matched)
	if def.Aggregation == "" {
		return result
	}
	aggregated := make(map[string]any)
	for i, output := range def.Outputs {
		var nums []float64
		for _, rule := range matched {
			if i < len(rule.Outputs) {
				if n, ok := toFloat64(rule.Outputs[i]); ok {
					nums = append(nums, n)
				}
			}
		}
		aggregated[output.Name] = aggregate(def.Aggregation, nums)
	}
	return entities.DecisionResult{Values: aggregated}
}

// buildResult maps output values from the first (or all) matched rules into the result.
func (e *DecisionTableEvaluatorImpl) buildResult(def entities.DecisionDefinition, rules []entities.DecisionRule) entities.DecisionResult {
	values := make(map[string]any)
	for _, rule := range rules {
		for i, output := range def.Outputs {
			if i < len(rule.Outputs) {
				values[output.Name] = rule.Outputs[i]
			}
		}
	}
	return entities.DecisionResult{Values: values}
}

// aggregate applies the aggregation function over a slice of numbers.
func aggregate(fn string, nums []float64) any {
	if len(nums) == 0 {
		return nil
	}
	switch fn {
	case entities.AggregationSum:
		var sum float64
		for _, n := range nums {
			sum += n
		}
		return sum
	case entities.AggregationCount:
		return float64(len(nums))
	case entities.AggregationMin:
		return slicesMin(nums)
	case entities.AggregationMax:
		return slicesMax(nums)
	}
	return nums
}

func slicesMin(nums []float64) float64 {
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}

func slicesMax(nums []float64) float64 {
	m := nums[0]
	for _, n := range nums[1:] {
		if n > m {
			m = n
		}
	}
	return m
}
