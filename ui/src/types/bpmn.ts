/**
 * Strict TypeScript types for BPMN node and edge data stored in React Flow.
 *
 * FE-ARCH-3: replaces all `any` casts for node.data and edge.data.
 *
 * Usage:
 *   import type { BPMNNodeData, BPMNEdgeData } from '../types/bpmn';
 *   const nodes: Node<BPMNNodeData>[] = [...];
 */

// ─── Shared base ─────────────────────────────────────────────────────────────

/** Properties every node type exposes regardless of specialisation. */
export interface BaseBPMNNodeData {
  label: string;
  documentation?: string;
  /** Raw property bag preserved from the server round-trip. */
  properties?: Record<string, unknown>;
  /** Set by the instance viewer to show execution state. */
  status?: 'active' | 'completed' | 'pending';
  /** Heatmap overlay value (execution frequency). */
  heatmapValue?: number;
  [key: string]: unknown;
}

/** Multi-instance configuration shared by several task types. */
export interface MultiInstanceConfig {
  multiInstanceType?: 'parallel' | 'sequential' | 'none';
  loopCardinality?: number;
  collection?: string;
  elementVariable?: string;
  completionCondition?: string;
}

// ─── Task variants ────────────────────────────────────────────────────────────

export interface UserTaskData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'userTask';
  assignee?: string;
  candidateUsers?: string[];
  candidateGroups?: string[];
  priority?: number;
  dueDate?: string;
  formKey?: string;
  formDefinition?: unknown;
}

export interface ManualTaskData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'manualTask';
  assignee?: string;
}

export interface ServiceTaskData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'serviceTask';
  /** External worker topic (pull model). */
  externalTopic?: string;
  implementation?: string;
  connector_instance_id?: string;
  lockDuration?: string;
  httpUrl?: string;
  httpMethod?: string;
  headers?: string;
  inputMapping?: string;
  outputMapping?: string;
  resultVariable?: string;
}

export interface ScriptTaskData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'scriptTask';
  script?: string;
  scriptFormat?: string;
}

export interface BusinessRuleTaskData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'businessRuleTask';
  decisionKey?: string;
  decisionVersion?: number;
  inputMapping?: string;
  outputMapping?: string;
}

export interface CallActivityData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'callActivity';
  calledElement?: string;
  calledElementVersion?: number;
}

// ─── Gateway variants ─────────────────────────────────────────────────────────

export interface GatewayData extends BaseBPMNNodeData {
  nodeType: 'exclusiveGateway' | 'inclusiveGateway' | 'parallelGateway' | 'eventBasedGateway';
  defaultFlow?: string;
}

// ─── Event variants ───────────────────────────────────────────────────────────

export interface EventData extends BaseBPMNNodeData {
  nodeType:
    | 'startEvent'
    | 'endEvent'
    | 'intermediateCatchEvent'
    | 'intermediateThrowEvent'
    | 'boundaryEvent'
    | 'signalEvent'
    | 'messageEvent'
    | 'timerEvent';
  eventType?: string;
  timerType?: 'duration' | 'date' | 'cycle';
  /** Timer duration or ISO date expression. Also used as the condition field. */
  duration?: string;
  signalName?: string;
  messageName?: string;
  correlationKey?: string;
  /** Boundary event: reference to the host activity. */
  attachedToRef?: string;
  /** Boundary event: whether the attached activity is cancelled. */
  cancelActivity?: boolean;
}

// ─── Sub-process ──────────────────────────────────────────────────────────────

export interface SubProcessData extends BaseBPMNNodeData, MultiInstanceConfig {
  nodeType: 'subProcess';
  isEventSubProcess?: boolean;
  parentId?: string;
}

// ─── Discriminated union ──────────────────────────────────────────────────────

export type BPMNNodeData =
  | UserTaskData
  | ManualTaskData
  | ServiceTaskData
  | ScriptTaskData
  | BusinessRuleTaskData
  | CallActivityData
  | GatewayData
  | EventData
  | SubProcessData;

// ─── Edge data ────────────────────────────────────────────────────────────────

export interface BPMNEdgeData {
  /** Condition expression displayed on the edge (used by gateways). */
  condition?: string;
  documentation?: string;
  [key: string]: unknown;
}

