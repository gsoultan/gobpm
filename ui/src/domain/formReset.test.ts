import { describe, expect, it } from 'bun:test';
import { formIdentity, shouldResetForm } from './formReset';

const fields = [{ id: 'amount', label: 'Amount', type: 'number' }];

describe('shouldResetForm', () => {
  it('does not reset when a refetch delivers the same content in new objects', () => {
    // The bug: `definition.fields || []` and `task.variables || {}` are new
    // objects every render, and identity comparison wiped the form each time.
    const before = formIdentity(fields, { region: 'EU' });
    const after = formIdentity([...fields.map((f) => ({ ...f }))], { region: 'EU' });
    expect(shouldResetForm(before, after)).toBe(false);
  });

  it('resets when the definition changes', () => {
    const before = formIdentity(fields, {});
    const after = formIdentity([...fields, { id: 'note', label: 'Note', type: 'text' }], {});
    expect(shouldResetForm(before, after)).toBe(true);
  });

  it('resets when the starting values change', () => {
    const before = formIdentity(fields, { amount: 1 });
    const after = formIdentity(fields, { amount: 2 });
    expect(shouldResetForm(before, after)).toBe(true);
  });
});
