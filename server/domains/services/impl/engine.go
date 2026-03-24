package impl

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/gsoultan/gobpm/server/domains/adapters"
	"github.com/gsoultan/gobpm/server/domains/entities"
	handlercontracts "github.com/gsoultan/gobpm/server/domains/handlers/contracts"
	"github.com/gsoultan/gobpm/server/domains/logic"
	observerContracts "github.com/gsoultan/gobpm/server/domains/observers/contracts"
	serviceContracts "github.com/gsoultan/gobpm/server/domains/services/contracts"
	"github.com/gsoultan/gobpm/server/repositories"
	"github.com/gsoultan/gobpm/server/repositories/models"
)

// Engine is the concrete BPMN execution engine.  It is exported so that the
// composition root (service.go) can call the wiring helpers (SetJobService etc.)
// without exposing those methods on the ExecutionEngine interface, satisfying ISP.
// All other consumers should depend on the serviceContracts.ExecutionEngine interface.
type Engine struct {
	repo           repositories.Repository
	handlerFactory handlercontracts.NodeHandlerFactory
	dispatcher     observerContracts.EventDispatcher
	jobSvc         serviceContracts.JobService
	varHistory     serviceContracts.VariableHistoryWriter
}

// EngineOption is a functional option for configuring an Engine after construction.
// Options are used by the composition root (service.go) to resolve circular
// dependencies (engine ↔ jobSvc, engine ↔ handlerFactory) without exposing
// mutable public setter methods on the Engine type.
type EngineOption func(*Engine)

// WithJobService injects the JobService used for timer and service-task enqueueing.
func WithJobService(js serviceContracts.JobService) EngineOption {
	return func(e *Engine) { e.jobSvc = js }
}

// WithHandlerFactory injects the factory used to resolve BPMN node handlers.
func WithHandlerFactory(hf handlercontracts.NodeHandlerFactory) EngineOption {
	return func(e *Engine) { e.handlerFactory = hf }
}

// WithVariableHistoryService injects the snapshot writer for variable history.
func WithVariableHistoryService(vh serviceContracts.VariableHistoryWriter) EngineOption {
	return func(e *Engine) { e.varHistory = vh }
}

// NewExecutionEngine constructs an Engine with its mandatory dependencies.
// Pass EngineOption values (WithJobService, WithHandlerFactory, etc.) to inject
// collaborators that depend on the engine itself (circular dependencies).
// Returns the concrete *Engine so the composition root can apply options; assign
// the result to serviceContracts.ExecutionEngine at the public API boundary.
func NewExecutionEngine(
	repo repositories.Repository,
	dispatcher observerContracts.EventDispatcher,
	opts ...EngineOption,
) *Engine {
	e := &Engine{repo: repo, dispatcher: dispatcher}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Apply runs additional EngineOptions on an already-constructed Engine.
// This is used by the composition root to inject collaborators that couldn't
// be provided at construction time due to circular dependencies.
func (e *Engine) Apply(opts ...EngineOption) {
	for _, opt := range opts {
		opt(e)
	}
}

func (e *Engine) StartProcess(ctx context.Context, projectID uuid.UUID, definitionKey string, vars map[string]any) (uuid.UUID, error) {
	return e.StartSubProcess(ctx, projectID, definitionKey, 0, vars, uuid.Nil, "")
}

func (e *Engine) StartSubProcess(ctx context.Context, projectID uuid.UUID, definitionKey string, version int, vars map[string]any, parentInstanceID uuid.UUID, parentNodeID string) (uuid.UUID, error) {
	var instanceID uuid.UUID
	err := e.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		cmd := NewStartProcessCommand(e, projectID, definitionKey, version, vars, parentInstanceID, parentNodeID)
		err := cmd.Execute(txCtx)
		instanceID = cmd.InstanceID
		return err
	})
	return instanceID, err
}

func (e *Engine) startProcessInternal(ctx context.Context, projectID uuid.UUID, definitionKey string, version int, vars map[string]any, parentInstanceID uuid.UUID, parentNodeID string) (uuid.UUID, error) {
	var m models.ProcessDefinitionModel
	var err error
	if version > 0 {
		m, err = e.repo.Definition().GetByKeyAndVersion(ctx, definitionKey, version)
	} else {
		m, err = e.repo.Definition().GetByKey(ctx, definitionKey)
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not find definition: %w", err)
	}
	def := adapters.DefinitionEntityAdapter{Model: m}.ToEntity()

	startNode := def.GetStartNode()
	if startNode == nil {
		return uuid.Nil, fmt.Errorf("definition %s has no start event", definitionKey)
	}

	idObj, _ := uuid.NewV7()

	instance := entities.ProcessInstance{
		ID:         idObj,
		Project:    &entities.Project{ID: projectID},
		Definition: &entities.ProcessDefinition{ID: def.ID},
		Status:     entities.ProcessActive,
		Variables:  vars,
		CreatedAt:  time.Now(),
	}
	instance.AddToken(startNode)

	if parentInstanceID != uuid.Nil {
		instance.ParentInstance = &entities.ProcessInstance{ID: parentInstanceID}
		instance.ParentNode = &entities.Node{ID: parentNodeID}
	}

	err = e.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		_, err := e.repo.Process().Create(txCtx, adapters.InstanceModelAdapter{Instance: instance}.ToModel())
		if err != nil {
			return err
		}

		// Activate Event Sub-processes start events for the process level
		for _, node := range def.Nodes {
			if node.IsEventSubProcess && node.ParentID == "" {
				// Find start event in this event sub-process
				for _, sn := range def.Nodes {
					if sn.ParentID == node.ID && sn.Type == entities.StartEvent {
						e.activateEventNode(txCtx, &instance, def, sn)
					}
				}
			}
		}

		e.dispatcher.Dispatch(txCtx, entities.ProcessEvent{
			Type:      entities.EventProcessStarted,
			Instance:  &instance,
			Project:   instance.Project,
			Timestamp: time.Now().Unix(),
			Variables: instance.Variables,
		})

		return e.ExecuteNode(txCtx, &instance, def, startNode.ID)
	})

	return idObj, err
}

func (e *Engine) GetInstance(ctx context.Context, id uuid.UUID) (entities.ProcessInstance, error) {
	m, err := e.repo.Process().Get(ctx, id)
	if err != nil {
		return entities.ProcessInstance{}, err
	}
	return adapters.InstanceEntityAdapter{Model: m}.ToEntity(), nil
}

func (e *Engine) GetInstanceForUpdate(ctx context.Context, id uuid.UUID) (entities.ProcessInstance, error) {
	m, err := e.repo.Process().GetForUpdate(ctx, id)
	if err != nil {
		return entities.ProcessInstance{}, err
	}
	return adapters.InstanceEntityAdapter{Model: m}.ToEntity(), nil
}

func (e *Engine) GetProcessDefinition(ctx context.Context, id uuid.UUID) (entities.ProcessDefinition, error) {
	m, err := e.repo.Definition().Get(ctx, id)
	if err != nil {
		return entities.ProcessDefinition{}, err
	}
	return adapters.DefinitionEntityAdapter{Model: m}.ToEntity(), nil
}

func (e *Engine) ListInstances(ctx context.Context, projectID uuid.UUID) ([]entities.ProcessInstance, error) {
	var ms []models.ProcessInstanceModel
	var err error
	if projectID != uuid.Nil {
		ms, err = e.repo.Process().ListByProject(ctx, projectID)
	} else {
		ms, err = e.repo.Process().List(ctx)
	}
	if err != nil {
		return nil, err
	}
	res := make([]entities.ProcessInstance, len(ms))
	for i, m := range ms {
		res[i] = adapters.InstanceEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (e *Engine) ListSubProcesses(ctx context.Context, parentInstanceID uuid.UUID) ([]entities.ProcessInstance, error) {
	ms, err := e.repo.Process().ListByParent(ctx, parentInstanceID)
	if err != nil {
		return nil, err
	}
	res := make([]entities.ProcessInstance, len(ms))
	for i, m := range ms {
		res[i] = adapters.InstanceEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

// GetRootInstance walks the parent chain starting from instanceID and returns
// the top-level ancestor. Stops if a cycle is detected (max 100 hops).
func (e *Engine) GetRootInstance(ctx context.Context, instanceID uuid.UUID) (entities.ProcessInstance, error) {
	const maxDepth = 100
	current, err := e.GetInstance(ctx, instanceID)
	if err != nil {
		return entities.ProcessInstance{}, fmt.Errorf("GetRootInstance: load instance: %w", err)
	}
	for depth := range maxDepth {
		if current.ParentInstance == nil {
			return current, nil
		}
		parent, err := e.GetInstance(ctx, current.ParentInstance.ID)
		if err != nil {
			return entities.ProcessInstance{}, fmt.Errorf("GetRootInstance: load parent at depth %d: %w", depth, err)
		}
		current = parent
	}
	return current, nil
}

func (e *Engine) GetExecutionPath(ctx context.Context, instanceID uuid.UUID) (entities.ExecutionPath, error) {
	entries, err := e.repo.Audit().ListByInstance(ctx, instanceID)
	if err != nil {
		return entities.ExecutionPath{}, err
	}

	var nodes []*entities.Node
	frequencies := make(map[string]int)
	seen := make(map[string]bool)

	// Audit logs are usually ordered by timestamp desc. We want chronological order.
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.Type == entities.EventNodeReached && entry.NodeID != "" {
			frequencies[entry.NodeID]++
			if !seen[entry.NodeID] {
				nodes = append(nodes, &entities.Node{ID: entry.NodeID})
				seen[entry.NodeID] = true
			}
		}
	}
	return entities.ExecutionPath{
		Nodes:       nodes,
		Frequencies: frequencies,
	}, nil
}

func (e *Engine) GetAuditLogs(ctx context.Context, instanceID uuid.UUID) ([]entities.AuditEntry, error) {
	ms, err := e.repo.Audit().ListByInstance(ctx, instanceID)
	if err != nil {
		return nil, err
	}
	res := make([]entities.AuditEntry, len(ms))
	for i, m := range ms {
		res[i] = adapters.AuditEntityAdapter{Model: m}.ToEntity()
	}
	return res, nil
}

func (e *Engine) ExecuteNode(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string) error {
	return e.ExecuteNodeIteration(ctx, instance, def, nodeID, "")
}

func (e *Engine) ExecuteNodeIteration(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string, iterationID string) error {
	return e.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		cmd := NewExecuteNodeCommand(e, instance, def, nodeID, iterationID)
		return cmd.Execute(txCtx)
	})
}

func (e *Engine) executeNodeInternal(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string, iterationID string) error {
	node := def.FindNode(nodeID)
	if node == nil {
		return fmt.Errorf("node %s not found", nodeID)
	}

	e.dispatcher.Dispatch(ctx, entities.ProcessEvent{
		Type:      entities.EventNodeReached,
		Instance:  instance,
		Project:   instance.Project,
		Node:      node,
		Timestamp: time.Now().Unix(),
		Variables: instance.Variables,
	})

	handler, err := e.handlerFactory.GetHandler(node.Type)
	if err != nil {
		return err
	}

	// Boundary Events: activation
	events := def.GetBoundaryEvents(node.ID)
	for _, event := range events {
		if signalName := event.GetStringProperty("signal_name"); signalName != "" {
			_ = e.repo.Subscription().Create(ctx, adapters.SubscriptionModelAdapter{Subscription: entities.NewSignalSubscription(instance.Project, instance, event, signalName)}.ToModel())
		}
		if messageName := event.GetStringProperty("message_name"); messageName != "" {
			correlationKey := event.GetStringProperty("correlation_key")
			_ = e.repo.Subscription().Create(ctx, adapters.SubscriptionModelAdapter{Subscription: entities.NewMessageSubscription(instance.Project, instance, event, messageName, correlationKey)}.ToModel())
		}
		if duration := event.GetStringProperty("timer_duration"); duration != "" {
			if e.jobSvc != nil {
				if err := e.jobSvc.EnqueueTimer(ctx, *instance, *event, duration); err != nil {
					return fmt.Errorf("failed to enqueue timer boundary: %w", err)
				}
			}
		}
	}

	err = handler.Execute(ctx, instance, def, *node, iterationID)
	if err != nil {
		bpmnError := ""
		if strings.HasPrefix(err.Error(), "BPMN_ERROR:") {
			bpmnError = strings.TrimPrefix(err.Error(), "BPMN_ERROR:")
		}

		// Boundary Events: check for error boundary events
		events := def.GetBoundaryEvents(node.ID)
		for _, event := range events {
			catchCode := event.GetStringProperty("error_code")
			if catchCode != "" {
				if bpmnError != "" && (catchCode == bpmnError || catchCode == "*") {
					return e.Proceed(ctx, instance, def, event.ID)
				}
				if bpmnError == "" && catchCode == "*" {
					return e.Proceed(ctx, instance, def, event.ID)
				}
			}
		}
		return err
	}

	return nil
}

func (e *Engine) Proceed(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string) error {
	return e.ProceedIteration(ctx, instance, def, nodeID, "")
}

func (e *Engine) ProceedIteration(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string, iterationID string) error {
	return e.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		cmd := NewProceedCommand(e, instance, def, nodeID, iterationID)
		return cmd.Execute(txCtx)
	})
}

// proceedInternal advances a process instance past nodeID at the end of the
// current transaction.  It is called by ProceedIteration via UnitOfWork.Do.
func (e *Engine) proceedInternal(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string, iterationID string) error {
	node := def.FindNode(nodeID)

	// Step 1: handle interrupting boundary events.
	e.handleBoundaryInterrupt(ctx, instance, def, node)

	// Step 2: remove the token (simple case) or check multi-instance completion.
	done, err := e.removeOrCheckMultiInstance(ctx, instance, def, node, nodeID, iterationID)
	if err != nil || !done {
		return err // not done yet — wait for remaining iterations
	}

	// Step 3: clean up sibling tokens and subscriptions.
	e.cleanupEventBasedGatewaySiblings(ctx, instance, def, nodeID)
	e.cleanupSubscriptions(ctx, instance, nodeID, def)

	// Step 4: mark node completed and follow outgoing flows.
	instance.MarkCompleted(node)
	return e.followOutgoingFlows(ctx, instance, def, nodeID)
}

// handleBoundaryInterrupt removes the host activity token when an interrupting
// boundary event fires and cleans up related subscriptions.
func (e *Engine) handleBoundaryInterrupt(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node *entities.Node) {
	if node == nil || node.Type != entities.BoundaryEvent || node.AttachedToRef == "" {
		return
	}
	hostNode := def.FindNode(node.AttachedToRef)
	if hostNode != nil {
		instance.RemoveTokenByNode(hostNode)
	}
	for _, ev := range def.GetBoundaryEvents(node.AttachedToRef) {
		_ = e.repo.Subscription().DeleteByNode(ctx, instance.ID, ev.ID)
	}
}

// removeOrCheckMultiInstance handles token removal for both simple and multi-instance
// nodes.  Returns (true, nil) when execution should continue past the node.
func (e *Engine) removeOrCheckMultiInstance(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node *entities.Node, nodeID, iterationID string) (bool, error) {
	if node == nil || node.MultiInstanceType == "" || node.MultiInstanceType == "none" {
		instance.RemoveTokenByNode(node)
		return true, nil
	}
	return e.checkMultiInstanceCompletion(ctx, instance, node, nodeID, iterationID)
}

// checkMultiInstanceCompletion increments the completion counter and returns
// (true, nil) when all iterations are done (or the completion condition is met).
func (e *Engine) checkMultiInstanceCompletion(ctx context.Context, instance *entities.ProcessInstance, node *entities.Node, nodeID, iterationID string) (bool, error) {
	completedKey := fmt.Sprintf("_mi_%s_completed", nodeID)
	totalKey := fmt.Sprintf("_mi_%s_total", nodeID)

	completed := extractInt(instance.Variables, completedKey) + 1
	total := extractInt(instance.Variables, totalKey)

	instance.SetVariable(completedKey, completed)
	instance.RemoveTokenByIteration(node, iterationID)

	conditionMet := completed >= total
	if node.CompletionCondition != "" {
		conditionMet = logic.GetConditionEvaluatorChain().Evaluate(node.CompletionCondition, instance.Variables)
	}

	if !conditionMet {
		return false, e.UpdateInstance(ctx, *instance)
	}

	// All iterations complete — clean up MI tracking variables.
	delete(instance.Variables, completedKey)
	delete(instance.Variables, totalKey)
	delete(instance.Variables, fmt.Sprintf("_mi_%s_active", nodeID))
	return true, nil
}

// cleanupEventBasedGatewaySiblings cancels competing tokens when one branch of
// an event-based gateway is taken.
func (e *Engine) cleanupEventBasedGatewaySiblings(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string) {
	for _, inFlow := range def.GetIncomingFlows(nodeID) {
		src := def.FindNode(inFlow.SourceRef)
		if src == nil || src.Type != entities.EventBasedGateway {
			continue
		}
		for _, siblingFlow := range def.GetOutgoingFlows(src.ID) {
			if siblingFlow.TargetRef != nodeID {
				targetNode := def.FindNode(siblingFlow.TargetRef)
				if targetNode != nil {
					instance.RemoveTokenByNode(targetNode)
				}
				_ = e.repo.Subscription().DeleteByNode(ctx, instance.ID, siblingFlow.TargetRef)
			}
		}
	}
}

// cleanupSubscriptions removes subscriptions for boundary events attached to
// nodeID and the catch-event subscription for nodeID itself.
func (e *Engine) cleanupSubscriptions(ctx context.Context, instance *entities.ProcessInstance, nodeID string, def entities.ProcessDefinition) {
	for _, ev := range def.GetBoundaryEvents(nodeID) {
		_ = e.repo.Subscription().DeleteByNode(ctx, instance.ID, ev.ID)
	}
	_ = e.repo.Subscription().DeleteByNode(ctx, instance.ID, nodeID)
}

// followOutgoingFlows adds tokens to every target node and executes them.
func (e *Engine) followOutgoingFlows(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID string) error {
	flows := def.GetOutgoingFlows(nodeID)
	if len(flows) == 0 {
		return e.UpdateInstance(ctx, *instance)
	}
	for _, flow := range flows {
		targetNode := def.FindNode(flow.TargetRef)
		if targetNode != nil {
			instance.AddToken(targetNode)
		}
		if err := e.UpdateInstance(ctx, *instance); err != nil {
			return err
		}
		if err := e.ExecuteNode(ctx, instance, def, flow.TargetRef); err != nil {
			return err
		}
	}
	return nil
}

// extractInt reads an integer process variable that may be stored as float64 (JSON default) or int.
func extractInt(vars map[string]any, key string) int {
	val, ok := vars[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func (e *Engine) activateEventNode(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node *entities.Node) {
	if signalName := node.GetStringProperty("signal_name"); signalName != "" {
		_ = e.repo.Subscription().Create(ctx, adapters.SubscriptionModelAdapter{Subscription: entities.NewSignalSubscription(instance.Project, instance, node, signalName)}.ToModel())
	}
	if messageName := node.GetStringProperty("message_name"); messageName != "" {
		correlationKey := node.GetStringProperty("correlation_key")
		_ = e.repo.Subscription().Create(ctx, adapters.SubscriptionModelAdapter{Subscription: entities.NewMessageSubscription(instance.Project, instance, node, messageName, correlationKey)}.ToModel())
	}
	if duration := node.GetStringProperty("timer_duration"); duration != "" {
		if e.jobSvc != nil {
			_ = e.jobSvc.EnqueueTimer(ctx, *instance, *node, duration)
		}
	}
}

func (e *Engine) UpdateInstance(ctx context.Context, instance entities.ProcessInstance) error {
	if err := e.repo.Process().Update(ctx, adapters.InstanceModelAdapter{Instance: instance}.ToModel()); err != nil {
		return err
	}
	e.captureVariableSnapshot(ctx, instance)
	return nil
}

// captureVariableSnapshot asynchronously writes a variable snapshot when a
// VariableHistoryWriter is configured. Errors are silently logged to avoid
// disrupting the main execution flow.
func (e *Engine) captureVariableSnapshot(ctx context.Context, instance entities.ProcessInstance) {
	if e.varHistory == nil {
		return
	}
	snap := entities.VariableSnapshot{
		Instance:   &instance,
		Variables:  maps.Clone(instance.Variables),
		CapturedAt: time.Now(),
	}
	_ = e.varHistory.CaptureSnapshot(ctx, snap)
}

func (e *Engine) DispatchEvent(ctx context.Context, event entities.ProcessEvent) {
	e.dispatcher.Dispatch(ctx, event)
}

func (e *Engine) BroadcastSignal(ctx context.Context, projectID uuid.UUID, signalName string, vars map[string]any) error {
	ms, err := e.repo.Subscription().FindSignals(ctx, projectID, signalName)
	if err != nil {
		return err
	}

	for _, m := range ms {
		e.triggerSubscription(ctx, adapters.SubscriptionEntityAdapter{Model: m}.ToEntity(), vars)
	}

	// Handle Signal Start Events
	return e.triggerStartEvents(ctx, projectID, "signal_name", signalName, vars)
}

func (e *Engine) SendMessage(ctx context.Context, projectID uuid.UUID, messageName, correlationKey string, vars map[string]any) error {
	ms, err := e.repo.Subscription().FindMessages(ctx, projectID, messageName, correlationKey)
	if err != nil {
		return err
	}

	for _, m := range ms {
		e.triggerSubscription(ctx, adapters.SubscriptionEntityAdapter{Model: m}.ToEntity(), vars)
	}

	// Handle Message Start Events
	if correlationKey == "" { // Message starts usually don't have correlation key
		return e.triggerStartEvents(ctx, projectID, "message_name", messageName, vars)
	}

	return nil
}

func (e *Engine) triggerSubscription(ctx context.Context, sub entities.EventSubscription, vars map[string]any) {
	if sub.Instance == nil {
		return
	}
	m, err := e.repo.Process().GetForUpdate(ctx, sub.Instance.ID)
	if err != nil {
		return
	}
	instance := adapters.InstanceEntityAdapter{Model: m}.ToEntity()

	if instance.Definition == nil {
		return
	}
	md, err := e.repo.Definition().Get(ctx, instance.Definition.ID)
	if err != nil {
		return
	}
	def := adapters.DefinitionEntityAdapter{Model: md}.ToEntity()

	// Update variables
	for k, v := range vars {
		instance.SetVariable(k, v)
	}

	_ = e.repo.UnitOfWork().Do(ctx, func(txCtx context.Context) error {
		_ = e.repo.Subscription().Delete(txCtx, sub.ID)
		return e.Proceed(txCtx, &instance, def, sub.Node.ID)
	})
}

func (e *Engine) triggerStartEvents(ctx context.Context, projectID uuid.UUID, propName, propValue string, vars map[string]any) error {
	ms, err := e.repo.Definition().ListByProject(ctx, projectID)
	if err != nil {
		return nil // Ignore error or handle accordingly
	}

	for _, m := range ms {
		def := adapters.DefinitionEntityAdapter{Model: m}.ToEntity()
		for _, node := range def.Nodes {
			if node.Type == entities.StartEvent && node.GetStringProperty(propName) == propValue {
				_, _ = e.StartProcess(ctx, projectID, def.Key, vars)
			}
		}
	}
	return nil
}

func (e *Engine) TriggerEscalation(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node entities.Node, escalationCode string) error {
	currNodeID := node.ID
	if node.ParentID != "" {
		currNodeID = node.ParentID
	}

	for currNodeID != "" {
		if err := e.checkBoundaryEscalation(ctx, instance, def, currNodeID, escalationCode); err == nil {
			return nil // Handled
		}

		if err := e.checkEventSubProcessEscalation(ctx, instance, def, currNodeID, escalationCode); err == nil {
			return nil // Handled
		}

		parent := def.FindNode(currNodeID)
		if parent != nil {
			currNodeID = parent.ParentID
		} else {
			currNodeID = ""
		}
	}
	return nil
}

func (e *Engine) checkBoundaryEscalation(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID, escalationCode string) error {
	boundaryEvents := def.GetBoundaryEvents(nodeID)
	for _, be := range boundaryEvents {
		if be.GetStringProperty("escalation_code") == escalationCode || be.GetStringProperty("escalation_code") == "" {
			return e.Proceed(ctx, instance, def, be.ID)
		}
	}
	return fmt.Errorf("no boundary escalation found")
}

func (e *Engine) checkEventSubProcessEscalation(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, nodeID, escalationCode string) error {
	var siblings []*entities.Node
	parent := def.FindNode(nodeID)
	if parent != nil {
		siblings = parent.Nodes
	} else {
		siblings = def.Nodes
	}

	for _, sib := range siblings {
		if !sib.IsEventSubProcess {
			continue
		}
		for _, sn := range sib.Nodes {
			if sn.Type == entities.StartEvent && sn.GetStringProperty("escalation_code") == escalationCode {
				instance.AddToken(sn)
				return e.ExecuteNode(ctx, instance, def, sn.ID)
			}
		}
	}
	return fmt.Errorf("no event sub-process escalation found")
}

func (e *Engine) TriggerCompensation(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node entities.Node, activityRef string) error {
	if activityRef != "" {
		refNode := def.FindNode(activityRef)
		if refNode != nil {
			_ = e.compensateActivity(ctx, instance, def, refNode)
		}
		return e.Proceed(ctx, instance, def, node.ID)
	}

	// Compensate all in reverse order
	for i := len(instance.CompletedNodes) - 1; i >= 0; i-- {
		n := instance.CompletedNodes[i]
		_ = e.compensateActivity(ctx, instance, def, n)
	}
	return e.Proceed(ctx, instance, def, node.ID)
}

func (e *Engine) compensateActivity(ctx context.Context, instance *entities.ProcessInstance, def entities.ProcessDefinition, node *entities.Node) error {
	if node == nil {
		return nil
	}
	if slices.Contains(instance.CompensatedNodes, node) {
		return nil
	}
	boundaryEvents := def.GetBoundaryEvents(node.ID)
	for _, be := range boundaryEvents {
		// A boundary event is a compensation event if it has a specific property or type.
		if be.GetStringProperty("event_type") == "compensation" || be.GetStringProperty("compensation") == "true" {
			instance.MarkCompensated(node)
			return e.Proceed(ctx, instance, def, be.ID)
		}
	}
	return nil
}

func (e *Engine) ExecuteScript(ctx context.Context, script string, scriptFormat string, variables map[string]any) (map[string]any, error) {
	if scriptFormat != "javascript" && scriptFormat != "" {
		return nil, fmt.Errorf("unsupported script format: %s", scriptFormat)
	}

	vm := goja.New()
	for k, v := range variables {
		vm.Set(k, v)
	}

	// Helper to set variables back
	updatedVars := maps.Clone(variables)

	vm.Set("setVar", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) >= 2 {
			name := call.Arguments[0].String()
			val := call.Arguments[1].Export()
			updatedVars[name] = val
		}
		return goja.Undefined()
	})

	_, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("script execution failed: %w", err)
	}

	// Sync variables modified in root scope
	for k := range variables {
		val := vm.Get(k)
		if val != nil {
			updatedVars[k] = val.Export()
		}
	}

	return updatedVars, nil
}
