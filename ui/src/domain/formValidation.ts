/**
 * Whether one form field's value is acceptable, and if not, why in words.
 *
 * Three things went wrong here before, all of them silent:
 *
 * A required `section` — or any field type the form cannot render — was still
 * validated, so Complete refused with no message and no field to fix, because
 * the field with the error was never on the screen.
 *
 * A `pattern` was compiled with `new RegExp` on every keystroke, and a
 * malformed one threw out of the validator and blanked the form.
 *
 * A `customJs` rule the bounded evaluator refused was logged to the console
 * and then treated as passing — the one outcome a validation rule must never
 * have.
 */
import { evaluateFormCondition, evaluateFormExpression, type FormScope } from './formExpression';

/** Field types the task form knows how to draw an input for. */
export const RENDERED_FIELD_TYPES = ['text', 'textarea', 'number', 'date', 'select', 'boolean'] as const;

/** Layout-only types: known, drawn, but never an input. */
export const LAYOUT_FIELD_TYPES = ['section'] as const;

export const UNSUPPORTED_FIELD_MESSAGE = "This field type isn't supported yet. It can be left as it is.";
export const REQUIRED_MESSAGE = 'This field is required';
export const INVALID_FORMAT_MESSAGE = 'Invalid format';

export function isRenderedFieldType(type: string): boolean {
  return (RENDERED_FIELD_TYPES as readonly string[]).includes(type);
}

export function isLayoutFieldType(type: string): boolean {
  return (LAYOUT_FIELD_TYPES as readonly string[]).includes(type);
}

/** The parts of a field definition that validation reads. */
export interface ValidatableField {
  id: string;
  type: string;
  required?: boolean;
  validation?: {
    pattern?: string;
    message?: string;
    /** A comparison in the form expression language; the name is historical. */
    customJs?: string;
  };
  logic?: {
    hiddenIf?: string;
  };
}

const patternCache = new Map<string, RegExp | null>();

/**
 * Compiles a pattern once. Null means the pattern is not a regular expression
 * — which is the form author's mistake, and must be said, not skipped.
 */
export function compilePattern(pattern: string): RegExp | null {
  const cached = patternCache.get(pattern);
  if (cached !== undefined) {
    return cached;
  }
  let compiled: RegExp | null;
  try {
    compiled = new RegExp(pattern);
  } catch {
    compiled = null;
  }
  patternCache.set(pattern, compiled);
  return compiled;
}

function ruleRefusal(reason: string): string {
  return `This field's rule could not be checked (${reason}). Ask whoever built the form to fix it.`;
}

function isEmpty(value: unknown): boolean {
  return !value && value !== 0 && value !== false;
}

/**
 * Validates one field against the current values.
 *
 * Returns the message to show, or null when the field is fine — or when it is
 * hidden, or not something the form draws an input for, since no message can
 * be acted on for a field that is not on the screen.
 */
export function validateFieldValue(field: ValidatableField, scope: FormScope): string | null {
  if (!isRenderedFieldType(field.type)) {
    return null;
  }
  if (evaluateFormCondition(field.logic?.hiddenIf || '', scope)) {
    return null;
  }

  const value = scope.data[field.id];
  if (field.required && isEmpty(value)) {
    return REQUIRED_MESSAGE;
  }
  if (isEmpty(value)) {
    return null;
  }

  return checkPattern(field, value) ?? checkRule(field, value, scope);
}

function checkPattern(field: ValidatableField, value: unknown): string | null {
  const pattern = field.validation?.pattern;
  if (!pattern) {
    return null;
  }
  const compiled = compilePattern(pattern);
  if (compiled === null) {
    return ruleRefusal('its format pattern is not valid');
  }
  return compiled.test(String(value)) ? null : field.validation?.message || INVALID_FORMAT_MESSAGE;
}

function checkRule(field: ValidatableField, value: unknown, scope: FormScope): string | null {
  const rule = field.validation?.customJs;
  if (!rule) {
    return null;
  }
  // `value` is exposed under `data.value` rather than as a bare name, because
  // the evaluator reads from `data` and `vars` and nothing else.
  const ruleScope: FormScope = { data: { ...scope.data, value }, vars: scope.vars };
  try {
    const outcome = evaluateFormExpression(rule, ruleScope);
    if (outcome === true) {
      return null;
    }
    return typeof outcome === 'string' ? outcome : field.validation?.message || 'Validation failed';
  } catch (error) {
    return ruleRefusal(error instanceof Error ? error.message : 'it is not a comparison');
  }
}
