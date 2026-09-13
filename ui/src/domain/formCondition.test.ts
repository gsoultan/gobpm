import { describe, expect, it } from 'bun:test';
import { evaluateFormCondition } from './formExpression';
import { buildFormCondition, formConditionRefusal, isTooRichForFormBuilder, parseFormCondition } from './formCondition';

describe('buildFormCondition', () => {
  it('reads the field from data, with the evaluator\'s equality operator', () => {
    expect(buildFormCondition({ fieldId: 'amount', operator: '==', value: '100' })).toBe('data.amount == 100');
    expect(buildFormCondition({ fieldId: 'status', operator: '!=', value: 'approved' })).toBe('data.status != "approved"');
    expect(buildFormCondition({ fieldId: 'ok', operator: '==', value: 'true' })).toBe('data.ok == true');
  });

  it('emits something the form evaluator actually accepts', () => {
    // The old builder wrote `amount = "x"`, which the evaluator refused and
    // treated as false — so the field was never hidden.
    const condition = buildFormCondition({ fieldId: 'mode', operator: '==', value: 'simple' });
    expect(formConditionRefusal(condition)).toBeNull();
    expect(evaluateFormCondition(condition, { data: { mode: 'simple' }, vars: {} })).toBe(true);
    expect(evaluateFormCondition(condition, { data: { mode: 'full' }, vars: {} })).toBe(false);
  });

  it('is empty until both sides are filled in', () => {
    expect(buildFormCondition({ fieldId: '', operator: '==', value: '1' })).toBe('');
    expect(buildFormCondition({ fieldId: 'a', operator: '==', value: ' ' })).toBe('');
  });

  it('does not double a user\'s own quotes, and survives a quote inside the value', () => {
    expect(buildFormCondition({ fieldId: 'a', operator: '==', value: "'x'" })).toBe('data.a == "x"');
    expect(buildFormCondition({ fieldId: 'a', operator: '==', value: '6" pipe' })).toBe("data.a == '6\" pipe'");
  });
});

describe('parseFormCondition', () => {
  it('reads back what the builder wrote', () => {
    expect(parseFormCondition('data.status != "approved"')).toEqual({ fieldId: 'status', operator: '!=', value: 'approved' });
    expect(parseFormCondition('data.amount >= 10')).toEqual({ fieldId: 'amount', operator: '>=', value: '10' });
  });

  it('opens the FEEL the gateway builder used to write here, so it can be fixed', () => {
    expect(parseFormCondition('amount = "x"')).toEqual({ fieldId: 'amount', operator: '==', value: 'x' });
  });

  it('refuses a compound expression rather than losing half of it', () => {
    expect(parseFormCondition('data.a == 1 && data.b == 2')).toBeNull();
    expect(isTooRichForFormBuilder('data.a == 1 || data.b')).toBe(true);
    expect(isTooRichForFormBuilder('')).toBe(false);
  });
});

describe('formConditionRefusal', () => {
  it('names the problem with the gateway-flavoured expression', () => {
    expect(formConditionRefusal('amount = "x"')).toMatch(/amount is not readable|= is not part/);
  });

  it('is silent for an empty or acceptable condition', () => {
    expect(formConditionRefusal('')).toBeNull();
    expect(formConditionRefusal('data.total > vars.limit')).toBeNull();
  });
});
