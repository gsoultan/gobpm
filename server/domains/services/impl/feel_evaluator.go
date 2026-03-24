package impl

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// FEELEvaluator is a Strategy implementation of ExpressionEvaluator that supports
// a practical subset of FEEL (Friendly Enough Expression Language) from the DMN spec.
// Supported syntax:
//   - Equality:   "value", == "value"
//   - Comparison: > n, < n, >= n, <= n, != "value"
//   - Range:      [1..10], (1..10], [1..10), (1..10)
//   - List:       "a","b","c"  (matches any element)
//   - Negation:   not("a"), not(1..10)
//   - Any-match:  - (wildcard, always true)
type FEELEvaluator struct{}

// NewFEELEvaluator creates a new FEELEvaluator.
func NewFEELEvaluator() *FEELEvaluator {
	return &FEELEvaluator{}
}

// Evaluate evaluates a FEEL expression against the given variables and returns the result.
func (e *FEELEvaluator) Evaluate(_ context.Context, expression string, variables map[string]any) (any, error) {
	expr := strings.TrimSpace(expression)
	if expr == "" || expr == "-" {
		return true, nil
	}
	// Variable reference: plain identifier
	if val, ok := variables[expr]; ok {
		return val, nil
	}
	// Try boolean literal
	if b, err := strconv.ParseBool(expr); err == nil {
		return b, nil
	}
	// Try numeric literal
	if n, err := strconv.ParseFloat(expr, 64); err == nil {
		return n, nil
	}
	// String literal: "value"
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return strings.Trim(expr, `"`), nil
	}
	return nil, fmt.Errorf("FEEL: unsupported expression %q", expr)
}

// EvaluateBool evaluates a FEEL condition expression against an input value stored in variables["_input"].
// It supports comparisons, ranges, lists, and negation as used in DMN rule cells.
func (e *FEELEvaluator) EvaluateBool(_ context.Context, expression string, variables map[string]any) (bool, error) {
	expr := strings.TrimSpace(expression)
	if expr == "" || expr == "-" {
		return true, nil
	}
	inputVal := variables["_input"]
	return e.matchCondition(expr, inputVal)
}

// matchCondition evaluates a single FEEL condition cell against the input value.
func (e *FEELEvaluator) matchCondition(expr string, inputVal any) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Negation: not(...)
	if strings.HasPrefix(expr, "not(") && strings.HasSuffix(expr, ")") {
		inner := expr[4 : len(expr)-1]
		matched, err := e.matchCondition(inner, inputVal)
		return !matched, err
	}

	// List: comma-separated values — match any
	if strings.Contains(expr, ",") {
		return e.matchList(expr, inputVal)
	}

	// Range: [a..b] or (a..b) etc.
	if e.isRange(expr) {
		return e.matchRange(expr, inputVal)
	}

	return e.matchScalar(expr, inputVal)
}

// matchList returns true if inputVal matches any element in the comma-separated list.
func (e *FEELEvaluator) matchList(expr string, inputVal any) (bool, error) {
	for item := range strings.SplitSeq(expr, ",") {
		matched, err := e.matchCondition(strings.TrimSpace(item), inputVal)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// isRange returns true if the expression looks like a FEEL range literal.
func (e *FEELEvaluator) isRange(expr string) bool {
	return (strings.HasPrefix(expr, "[") || strings.HasPrefix(expr, "(")) &&
		strings.Contains(expr, "..") &&
		(strings.HasSuffix(expr, "]") || strings.HasSuffix(expr, ")"))
}

// matchRange evaluates a FEEL range expression against a numeric input value.
func (e *FEELEvaluator) matchRange(expr string, inputVal any) (bool, error) {
	inclusiveMin := strings.HasPrefix(expr, "[")
	inclusiveMax := strings.HasSuffix(expr, "]")
	inner := expr[1 : len(expr)-1]

	parts := strings.SplitN(inner, "..", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("FEEL: invalid range %q", expr)
	}

	inputNum, ok := toFloat64(inputVal)
	if !ok {
		return false, nil
	}
	minVal, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return false, fmt.Errorf("FEEL: invalid range min in %q", expr)
	}
	maxVal, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return false, fmt.Errorf("FEEL: invalid range max in %q", expr)
	}

	minOk := inclusiveMin && inputNum >= minVal || !inclusiveMin && inputNum > minVal
	maxOk := inclusiveMax && inputNum <= maxVal || !inclusiveMax && inputNum < maxVal
	return minOk && maxOk, nil
}

// matchScalar evaluates a comparison or equality expression against the input value.
func (e *FEELEvaluator) matchScalar(expr string, inputVal any) (bool, error) {
	for _, op := range []string{">=", "<=", "!=", ">", "<", "=="} {
		if strings.HasPrefix(expr, op) {
			return e.matchComparison(op, strings.TrimSpace(expr[len(op):]), inputVal)
		}
	}
	// Plain equality
	return e.matchEquality(expr, inputVal), nil
}

// matchComparison evaluates an operator-prefixed FEEL condition.
func (e *FEELEvaluator) matchComparison(op, rhs string, inputVal any) (bool, error) {
	rhsClean := strings.Trim(rhs, `"'`)
	inputStr := fmt.Sprintf("%v", inputVal)

	inputNum, inputIsNum := toFloat64(inputVal)
	rhsNum, rhsIsNum := parseNumber(rhsClean)

	switch op {
	case "==":
		return inputStr == rhsClean, nil
	case "!=":
		return inputStr != rhsClean, nil
	case ">":
		if inputIsNum && rhsIsNum {
			return inputNum > rhsNum, nil
		}
		return inputStr > rhsClean, nil
	case "<":
		if inputIsNum && rhsIsNum {
			return inputNum < rhsNum, nil
		}
		return inputStr < rhsClean, nil
	case ">=":
		if inputIsNum && rhsIsNum {
			return inputNum >= rhsNum, nil
		}
		return inputStr >= rhsClean, nil
	case "<=":
		if inputIsNum && rhsIsNum {
			return inputNum <= rhsNum, nil
		}
		return inputStr <= rhsClean, nil
	}
	return false, fmt.Errorf("FEEL: unknown operator %q", op)
}

// matchEquality checks plain value equality between the expression and the input.
func (e *FEELEvaluator) matchEquality(expr string, inputVal any) bool {
	clean := strings.Trim(expr, `"'`)
	return reflect.DeepEqual(fmt.Sprintf("%v", inputVal), clean)
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

func parseNumber(s string) (float64, bool) {
	n, err := strconv.ParseFloat(s, 64)
	return n, err == nil
}
