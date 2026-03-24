package impl

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/adapters"
	"github.com/gsoultan/gobpm/server/domains/entities"
	servicecontracts "github.com/gsoultan/gobpm/server/domains/services/contracts"
	"github.com/gsoultan/gobpm/server/repositories"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

type decisionService struct {
	repo           repositories.Repository
	tableEvaluator servicecontracts.DecisionTableEvaluator
}

// NewDecisionService creates a new DecisionService implementation.
// tableEvaluator is the Strategy used to evaluate decision table rules and hit policies.
func NewDecisionService(repo repositories.Repository, tableEvaluator servicecontracts.DecisionTableEvaluator) servicecontracts.DecisionService {
	return &decisionService{repo: repo, tableEvaluator: tableEvaluator}
}

func (s *decisionService) Evaluate(ctx context.Context, decisionKey string, version int, variables map[string]any) (entities.DecisionResult, error) {
	// Use a copy of variables to avoid polluting caller's map during intermediate steps
	varsCopy := make(map[string]any)
	for k, v := range variables {
		varsCopy[k] = v
	}
	return s.evaluateRecursive(ctx, decisionKey, version, varsCopy, make(map[string]bool))
}

func (s *decisionService) evaluateRecursive(ctx context.Context, decisionKey string, version int, variables map[string]any, seen map[string]bool) (entities.DecisionResult, error) {
	if seen[decisionKey] {
		return entities.DecisionResult{}, fmt.Errorf("circular dependency detected for decision %s", decisionKey)
	}
	seen[decisionKey] = true
	defer delete(seen, decisionKey)

	var m models.DecisionDefinitionModel
	var err error
	if version > 0 {
		m, err = s.repo.Decision().GetByKeyAndVersion(ctx, decisionKey, version)
	} else {
		m, err = s.repo.Decision().GetByKey(ctx, decisionKey)
	}
	if err != nil {
		return entities.DecisionResult{}, err
	}

	decision := adapters.DecisionEntityAdapter{Model: m}.ToEntity()

	// 1. Evaluate required decisions
	for _, reqKey := range decision.RequiredDecisions {
		res, err := s.evaluateRecursive(ctx, reqKey, 0, variables, seen)
		if err != nil {
			return entities.DecisionResult{}, fmt.Errorf("failed to evaluate required decision %s: %w", reqKey, err)
		}
		// DMN says it should be available as decision name/key.
		// If decision has multiple outputs, use a map. If single output, use it directly.
		if len(res.Values) == 1 {
			for _, v := range res.Values {
				variables[reqKey] = v
				break
			}
		} else {
			variables[reqKey] = res.Values
		}
	}

	// 2. Evaluate rules and apply hit policy via the injected Strategy
	return s.tableEvaluator.EvaluateTable(ctx, decision, variables)
}

func (s *decisionService) applyHitPolicy(decision entities.DecisionDefinition, matchingRules []entities.DecisionRule) (entities.DecisionResult, error) {
	hp := strings.ToUpper(decision.HitPolicy)
	if hp == "" {
		hp = entities.HitPolicyFirst
	}

	switch hp {
	case entities.HitPolicyUnique:
		if len(matchingRules) > 1 {
			return entities.DecisionResult{}, fmt.Errorf("UNIQUE hit policy violated: multiple rules matched")
		}
		return s.buildResult(decision, matchingRules[0]), nil

	case entities.HitPolicyFirst:
		return s.buildResult(decision, matchingRules[0]), nil

	case entities.HitPolicyAny:
		firstResult := s.buildResult(decision, matchingRules[0])
		for i := 1; i < len(matchingRules); i++ {
			nextResult := s.buildResult(decision, matchingRules[i])
			if !s.resultsEqual(firstResult, nextResult) {
				return entities.DecisionResult{}, fmt.Errorf("ANY hit policy violated: matching rules have different outputs")
			}
		}
		return firstResult, nil

	case entities.HitPolicyCollect:
		if decision.Aggregation == "" {
			var list []any
			for _, rule := range matchingRules {
				res := s.buildResult(decision, rule).Values
				if len(decision.Outputs) == 1 {
					for _, v := range res {
						list = append(list, v)
						break
					}
				} else {
					list = append(list, res)
				}
			}
			if len(decision.Outputs) == 1 {
				return entities.DecisionResult{Values: map[string]any{decision.Outputs[0].Name: list}}, nil
			}
			return entities.DecisionResult{Values: map[string]any{"results": list}}, nil
		}
		return s.applyAggregation(decision, matchingRules)

	case entities.HitPolicyPriority:
		// Simple implementation of priority: use first for now as we don't have output priority values
		return s.buildResult(decision, matchingRules[0]), nil

	default:
		return s.buildResult(decision, matchingRules[0]), nil
	}
}

func (s *decisionService) applyAggregation(decision entities.DecisionDefinition, matchingRules []entities.DecisionRule) (entities.DecisionResult, error) {
	if len(decision.Outputs) != 1 {
		return entities.DecisionResult{}, fmt.Errorf("aggregation only supported for decisions with exactly one output")
	}

	outputName := decision.Outputs[0].Name
	var values []float64
	for _, rule := range matchingRules {
		if len(rule.Outputs) > 0 {
			switch v := rule.Outputs[0].(type) {
			case float64:
				values = append(values, v)
			case int:
				values = append(values, float64(v))
			case int64:
				values = append(values, float64(v))
			case float32:
				values = append(values, float64(v))
			}
		}
	}

	if len(values) == 0 && strings.ToUpper(decision.Aggregation) != entities.AggregationCount {
		return entities.DecisionResult{}, fmt.Errorf("no numeric values for aggregation %s", decision.Aggregation)
	}

	var result float64
	switch strings.ToUpper(decision.Aggregation) {
	case entities.AggregationSum:
		for _, v := range values {
			result += v
		}
	case entities.AggregationCount:
		result = float64(len(values))
	case entities.AggregationMin:
		result = values[0]
		for _, v := range values {
			if v < result {
				result = v
			}
		}
	case entities.AggregationMax:
		result = values[0]
		for _, v := range values {
			if v > result {
				result = v
			}
		}
	default:
		return entities.DecisionResult{}, fmt.Errorf("unsupported aggregation: %s", decision.Aggregation)
	}

	return entities.DecisionResult{Values: map[string]any{outputName: result}}, nil
}

func (s *decisionService) resultsEqual(r1, r2 entities.DecisionResult) bool {
	return reflect.DeepEqual(r1.Values, r2.Values)
}

func (s *decisionService) buildResult(decision entities.DecisionDefinition, rule entities.DecisionRule) entities.DecisionResult {
	result := make(map[string]any)
	for i, outputVal := range rule.Outputs {
		if i >= len(decision.Outputs) {
			break
		}
		outputDef := decision.Outputs[i]
		result[outputDef.Name] = outputVal
	}
	return entities.DecisionResult{Values: result}
}

func (s *decisionService) ListDecisions(ctx context.Context, projectID uuid.UUID) ([]entities.DecisionDefinition, error) {
	var ms []models.DecisionDefinitionModel
	var err error
	if projectID != uuid.Nil {
		ms, err = s.repo.Decision().ListByProject(ctx, projectID)
	} else {
		ms, err = s.repo.Decision().List(ctx)
	}
	if err != nil {
		return nil, err
	}
	res := make([]entities.DecisionDefinition, len(ms))
	for i, m := range ms {
		res[i] = adapters.DecisionEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (s *decisionService) GetDecision(ctx context.Context, id uuid.UUID) (entities.DecisionDefinition, error) {
	m, err := s.repo.Decision().Get(ctx, id)
	if err != nil {
		return entities.DecisionDefinition{}, err
	}
	return adapters.DecisionEntityAdapter{Model: m}.ToEntity(), nil
}

func (s *decisionService) CreateDecision(ctx context.Context, d entities.DecisionDefinition) (uuid.UUID, error) {
	err := s.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		if d.ID == uuid.Nil {
			d.ID, _ = uuid.NewV7()
		}

		// Increment version if key already exists
		existing, err := s.repo.Decision().GetByKey(txCtx, d.Key)
		if err == nil {
			d.Version = existing.Version + 1
		} else {
			d.Version = 1
		}

		return s.repo.Decision().Create(txCtx, adapters.DecisionModelAdapter{Decision: d}.ToModel())
	})

	return d.ID, err
}

func (s *decisionService) UpdateDecision(ctx context.Context, id uuid.UUID, d entities.DecisionDefinition) error {
	d.ID = id
	return s.repo.Decision().Update(ctx, id, adapters.DecisionModelAdapter{Decision: d}.ToModel())
}

func (s *decisionService) DeleteDecision(ctx context.Context, id uuid.UUID) error {
	return s.repo.Decision().Delete(ctx, id)
}

func (s *decisionService) evaluateRule(rule entities.DecisionRule, inputDefs []entities.DecisionInput, variables map[string]any) (bool, error) {
	vm := goja.New()
	for k, v := range variables {
		vm.Set(k, v)
	}

	for i, cellExpr := range rule.Inputs {
		if i >= len(inputDefs) {
			break
		}
		if cellExpr == "" || cellExpr == "-" {
			continue // ignore cell
		}

		inputDef := inputDefs[i]
		val, ok := variables[inputDef.Expression]
		if !ok {
			// Try to evaluate the expression if it's not a direct variable name
			res, err := vm.RunString(inputDef.Expression)
			if err != nil {
				return false, nil // Assume no match on evaluation error for now
			}
			val = res.Export()
		}

		if !s.evaluateFeel(val, cellExpr, vm) {
			return false, nil
		}
	}

	return true, nil
}

func (s *decisionService) evaluateFeel(val any, cellExpr string, vm *goja.Runtime) bool {
	trimmed := strings.TrimSpace(cellExpr)
	if trimmed == "" || trimmed == "-" {
		return true
	}

	// Handle List: [val1, val2, val3]
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") && !strings.Contains(trimmed, "..") {
		content := trimmed[1 : len(trimmed)-1]
		parts := strings.Split(content, ",")
		for _, p := range parts {
			if s.evaluateFeel(val, strings.TrimSpace(p), vm) {
				return true
			}
		}
		return false
	}

	// Handle Range: [min..max], (min..max), [min..max), (min..max]
	if (strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "(")) && strings.Contains(trimmed, "..") {
		return s.evaluateRange(val, trimmed)
	}

	vm.Set("cellInput", val)

	// Check if cellExpr starts with an operator. If not, default to equality check
	expression := cellExpr
	hasOperator := false
	for _, op := range []string{"==", "!=", ">", "<", ">=", "<="} {
		if strings.HasPrefix(trimmed, op) {
			hasOperator = true
			break
		}
	}

	if !hasOperator {
		// If it's a string that doesn't look like a number or boolean or quoted string, quote it
		isLiteral := true
		if trimmed == "true" || trimmed == "false" || trimmed == "null" {
			isLiteral = false
		} else {
			// check if it's a number
			isLiteral = !s.isNumeric(trimmed)
		}
		if isLiteral && !strings.HasPrefix(trimmed, "'") && !strings.HasPrefix(trimmed, "\"") {
			expression = fmt.Sprintf("cellInput == '%s'", trimmed)
		} else {
			expression = fmt.Sprintf("cellInput == %s", trimmed)
		}
	} else {
		expression = fmt.Sprintf("cellInput %s", trimmed)
	}

	res, err := vm.RunString(expression)
	if err != nil {
		return false // assume no match
	}

	return res.ToBoolean()
}

func (s *decisionService) evaluateRange(val any, rangeExpr string) bool {
	// Simple parser for [min..max]
	trimmed := strings.TrimSpace(rangeExpr)
	opening := trimmed[0]
	closing := trimmed[len(trimmed)-1]
	content := trimmed[1 : len(trimmed)-1]
	parts := strings.Split(content, "..")
	if len(parts) != 2 {
		return false
	}

	minStr := strings.TrimSpace(parts[0])
	maxStr := strings.TrimSpace(parts[1])

	var min, max, v float64
	fmt.Sscanf(minStr, "%f", &min)
	fmt.Sscanf(maxStr, "%f", &max)

	switch valT := val.(type) {
	case float64:
		v = valT
	case int:
		v = float64(valT)
	case int64:
		v = float64(valT)
	case float32:
		v = float64(valT)
	default:
		return false
	}

	matchMin := false
	if opening == '[' {
		matchMin = v >= min
	} else {
		matchMin = v > min
	}

	matchMax := false
	if closing == ']' {
		matchMax = v <= max
	} else {
		matchMax = v < max
	}

	return matchMin && matchMax
}

func (s *decisionService) isNumeric(s_ string) bool {
	var f float64
	_, err := fmt.Sscanf(s_, "%f", &f)
	return err == nil
}
