import { describe, expect, it } from 'bun:test';
import { parseFormDefinition } from './formDefinition';

describe('parseFormDefinition', () => {
  it('reads a list of fields', () => {
    const parsed = parseFormDefinition<{ id: string }>('[{"id":"amount"}]');
    expect(parsed.fields.map((f) => f.id)).toEqual(['amount']);
    expect(parsed.error).toBeNull();
  });

  it('reads the wrapped shape too', () => {
    expect(parseFormDefinition('{"fields":[{"id":"a"}]}').fields).toHaveLength(1);
  });

  it('treats no definition as a form with no inputs', () => {
    expect(parseFormDefinition(undefined)).toEqual({ fields: [], error: null });
    expect(parseFormDefinition('  ')).toEqual({ fields: [], error: null });
  });

  it('refuses a definition it cannot read instead of rendering an empty form', () => {
    // An empty form shows "No inputs required" and a Complete button, which
    // is how an unreadable form was completing tasks with nothing filled in.
    expect(parseFormDefinition('{not json').error).toContain('could not be read');
    expect(parseFormDefinition('{"x":1}').error).toContain('not a list of fields');
  });
});
