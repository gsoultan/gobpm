/**
 * A decision, as a sentence.
 *
 * The timeline rendered a decision as `expense_approval_level v3`, `rule_1`
 * and `{"approval_level":"director"}` — the audit record, verbatim. The person
 * reading a timeline wants to know what was decided and by which policy:
 * "Decided by Expense approval level: Approval level: director".
 */
import { humanizeIdentifier } from './wording';

export interface DecisionNarrative {
  /** "Decided by Expense approval level: Approval level: director" */
  sentence: string;
  /** "Policy version 3", or null when the version is not recorded. */
  version: string | null;
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === 'object' && !Array.isArray(value) ? (value as Record<string, unknown>) : null;
}

function describeValue(value: unknown): string {
  if (value === null || value === undefined) return 'nothing';
  if (typeof value === 'string') return value;
  if (typeof value === 'number' || typeof value === 'boolean') return String(value);
  if (Array.isArray(value)) return value.map(describeValue).join(', ');
  const record = asRecord(value);
  if (record) return describeOutputs(record);
  return String(value);
}

function describeOutputs(outputs: Record<string, unknown>): string {
  const parts = Object.entries(outputs).map(([key, value]) => `${humanizeIdentifier(key)}: ${describeValue(value)}`);
  return parts.length > 0 ? parts.join(', ') : 'no outcome';
}

/** Which policy decided: its display name, or its key turned into words. */
function policyName(data: Record<string, unknown>): string {
  const name = data.decision_name;
  if (typeof name === 'string' && name.trim() !== '') return name;
  const key = data.decision_key;
  if (typeof key === 'string' && key.trim() !== '') return humanizeIdentifier(key);
  return 'a decision table';
}

export function describeDecision(data: Record<string, unknown> | undefined): DecisionNarrative {
  const record = data ?? {};
  const outcome = record.outputs === undefined ? 'no outcome recorded' : describeValue(record.outputs);
  const version = record.decision_version;
  return {
    sentence: `Decided by ${policyName(record)}: ${outcome}`,
    version: typeof version === 'number' || (typeof version === 'string' && version !== '')
      ? `Policy version ${version}`
      : null,
  };
}
