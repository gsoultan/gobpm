/**
 * The comparison a form field's Logic tab edits as three fields:
 * field · operator · value — written in the language the form evaluator reads.
 *
 * The Logic tab used the gateway builder, which emits FEEL (`amount = "x"`).
 * Form logic is evaluated in the browser by `formExpression`, which has no `=`
 * operator and reads only from `data` and `vars`, so every condition built in
 * the tab was refused at runtime and the field was never hidden. Nothing said
 * so.
 *
 * This builder emits `data.<fieldId> == <literal>`, reads back what it wrote
 * (and the old FEEL shape, so existing definitions open for editing), and
 * asks the evaluator whether the result is acceptable so the builder can show
 * a refusal rather than a green preview over a rule that does nothing.
 */
import type { ComparisonOperator } from './conditionExpression';
import { evaluateFormExpression } from './formExpression';

export interface FormComparison {
  /** The form field being compared; `data.` is added when the condition is built. */
  fieldId: string;
  operator: ComparisonOperator;
  /** As the user typed it; quoting is decided when the condition is built. */
  value: string;
}

export const FORM_OPERATORS: ComparisonOperator[] = ['==', '!=', '>', '<', '>=', '<='];

function stripQuotes(value: string): string {
  if (value.length >= 2) {
    const first = value[0];
    if ((first === '"' || first === "'") && value.endsWith(first)) {
      return value.slice(1, -1);
    }
  }
  return value;
}

/**
 * Numbers and booleans stay bare; anything else is quoted, or `status !=
 * approved` would try to read a second field. The evaluator has no escape
 * handling, so a value containing double quotes is wrapped in single ones.
 */
function toLiteral(value: string): string {
  const trimmed = value.trim();
  if (['true', 'false', 'null'].includes(trimmed)) {
    return trimmed;
  }
  if (trimmed !== '' && !Number.isNaN(Number(trimmed))) {
    return trimmed;
  }
  const bare = stripQuotes(trimmed);
  return bare.includes('"') ? `'${bare}'` : `"${bare}"`;
}

/** Builds the stored condition; empty until both the field and the value are filled in. */
export function buildFormCondition(comparison: FormComparison): string {
  const fieldId = comparison.fieldId.trim();
  const value = comparison.value.trim();
  if (fieldId === '' || value === '') {
    return '';
  }
  return `data.${fieldId} ${comparison.operator} ${toLiteral(value)}`;
}

const SHAPE = /^(?:data\.)?([A-Za-z_$][\w$]*)\s*(==|!=|>=|<=|=|>|<)\s*(.+)$/;

/**
 * Reads a stored condition back into the three fields. Accepts what this
 * builder writes and the FEEL the gateway builder used to write here; returns
 * null for anything richer, so an edit cannot silently overwrite it.
 */
export function parseFormCondition(condition: string): FormComparison | null {
  const match = SHAPE.exec(condition.trim());
  if (!match) {
    return null;
  }
  const [, fieldId, operator, rest] = match;
  const withoutStrings = rest.replace(/"[^"]*"|'[^']*'/g, '""');
  if (/[<>!=&|]/.test(withoutStrings)) {
    return null;
  }
  return {
    fieldId,
    operator: operator === '=' ? '==' : (operator as ComparisonOperator),
    value: stripQuotes(rest.trim()),
  };
}

/** True when a condition exists but is more than three fields can hold. */
export function isTooRichForFormBuilder(condition: string): boolean {
  return condition.trim() !== '' && parseFormCondition(condition) === null;
}

/**
 * Why the form evaluator would refuse this condition, in words — or null when
 * it would run. Evaluated against empty data: a rule that reads a missing
 * field is fine, a rule the grammar does not cover is not.
 */
export function formConditionRefusal(condition: string): string | null {
  if (condition.trim() === '') {
    return null;
  }
  try {
    evaluateFormExpression(condition, { data: {}, vars: {} });
    return null;
  } catch (error) {
    return error instanceof Error ? error.message : 'the rule is not a comparison';
  }
}
