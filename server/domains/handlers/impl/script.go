package impl

import (
	"context"
	"fmt"

	"github.com/dop251/goja"
	"github.com/gsoultan/gobpm/server/domains/entities"
	servicecontracts "github.com/gsoultan/gobpm/server/domains/services/contracts"
)

type ScriptTaskHandler struct {
	engine servicecontracts.EngineRunner
}

func (h *ScriptTaskHandler) DoExecute(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node entities.Node, iterationID string) error {
	script := node.Script
	if script == "" {
		script = node.Condition
	}
	if s, ok := node.Properties["script"].(string); ok && script == "" {
		script = s
	}

	if script == "" {
		return h.engine.ProceedIteration(ctx, instance, def, node.ID, iterationID)
	}

	vm := goja.New()

	// Expose variables to JS
	for k, v := range instance.Variables {
		vm.Set(k, v)
	}

	// Helper to set variables back
	vm.Set("setVar", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) >= 2 {
			name := call.Arguments[0].String()
			val := call.Arguments[1].Export()
			instance.SetVariable(name, val)
		}
		return goja.Undefined()
	})

	_, err := vm.RunString(script)
	if err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	// Also sync all variables that were modified in the root scope
	for k := range instance.Variables {
		val := vm.Get(k)
		if val != nil {
			instance.SetVariable(k, val.Export())
		}
	}

	return h.engine.ProceedIteration(ctx, instance, def, node.ID, iterationID)
}
