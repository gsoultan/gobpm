import { describe, expect, it } from 'bun:test';
import { compilePattern, REQUIRED_MESSAGE, validateFieldValue } from './formValidation';

const scope = (data: Record<string, unknown>) => ({ data, vars: {} });

describe('validateFieldValue', () => {
  it('never blocks on a field the form does not draw', () => {
    // A required section, or a type this version cannot render, used to fail
    // validation with no message and nothing on screen to fix.
    expect(validateFieldValue({ id: 's', type: 'section', required: true }, scope({}))).toBeNull();
    expect(validateFieldValue({ id: 'f', type: 'file', required: true }, scope({}))).toBeNull();
  });

  it('requires a value, but accepts zero and false', () => {
    expect(validateFieldValue({ id: 'n', type: 'number', required: true }, scope({}))).toBe(REQUIRED_MESSAGE);
    expect(validateFieldValue({ id: 'n', type: 'number', required: true }, scope({ n: 0 }))).toBeNull();
    expect(validateFieldValue({ id: 'b', type: 'boolean', required: true }, scope({ b: false }))).toBeNull();
  });

  it('checks a pattern and uses the author\'s message', () => {
    const field = { id: 'code', type: 'text', validation: { pattern: '^[0-9]+$', message: 'Digits only' } };
    expect(validateFieldValue(field, scope({ code: 'abc' }))).toBe('Digits only');
    expect(validateFieldValue(field, scope({ code: '123' }))).toBeNull();
  });

  it('says so when a pattern is not a regular expression, instead of throwing', () => {
    const field = { id: 'code', type: 'text', validation: { pattern: '([' } };
    expect(validateFieldValue(field, scope({ code: 'x' }))).toContain('could not be checked');
  });

  it('surfaces a refused rule as a visible failure, never as a pass', () => {
    const field = { id: 'amount', type: 'number', validation: { customJs: 'value => value > 10' } };
    expect(validateFieldValue(field, scope({ amount: 500 }))).toContain('could not be checked');
  });

  it('applies a rule that is a comparison', () => {
    const field = { id: 'amount', type: 'number', validation: { customJs: 'data.value > 10', message: 'Too small' } };
    expect(validateFieldValue(field, scope({ amount: 5 }))).toBe('Too small');
    expect(validateFieldValue(field, scope({ amount: 50 }))).toBeNull();
  });

  it('skips a hidden field entirely', () => {
    const field = { id: 'x', type: 'text', required: true, logic: { hiddenIf: 'data.mode == "simple"' } };
    expect(validateFieldValue(field, scope({ mode: 'simple' }))).toBeNull();
    expect(validateFieldValue(field, scope({ mode: 'full' }))).toBe(REQUIRED_MESSAGE);
  });
});

describe('compilePattern', () => {
  it('compiles a pattern once and hands back the same object', () => {
    expect(compilePattern('^a$')).toBe(compilePattern('^a$'));
  });

  it('returns null for a pattern that cannot compile', () => {
    expect(compilePattern('(')).toBeNull();
  });
});
