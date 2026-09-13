import { describe, expect, it } from 'bun:test';

import { shortCode, taskReference } from './taskReference';

describe('a short code that actually distinguishes', () => {
  /*
   * The reason this exists. Identifiers are UUIDv7, so the leading characters
   * encode the time: three approvals created in the same second all showed
   * `01a07764` in the inbox. Taking the end gives characters that differ.
   */
  it('differs between instances created in the same period', () => {
    const a = shortCode('01a07764-6839-75b1-8ffe-ea705bb67a3b');
    const b = shortCode('01a07764-6839-75b1-8ffe-ea705bb67c91');
    expect(a).not.toBe(b);
  });

  it('is short, upper case and free of separators', () => {
    expect(shortCode('01a07764-6839-75b1-8ffe-ea705bb67a3b')).toBe('B67A3B');
  });

  it('answers empty for a missing identifier rather than inventing one', () => {
    expect(shortCode(undefined)).toBe('');
    expect(shortCode('')).toBe('');
  });

  it('copes with an identifier shorter than the code', () => {
    expect(shortCode('abc')).toBe('ABC');
  });
});

describe('saying which piece of work this is', () => {
  it('uses the description the process was started with', () => {
    const { label } = taskReference(
      { amount: 1750, currency: 'GBP', description: 'Expense of GBP 1,750' },
      '01a07764-6839-75b1-8ffe-ea705bb67a3b',
    );
    expect(label).toBe('Expense of GBP 1,750');
  });

  it('prefers an explicit reference over a description', () => {
    const { label } = taskReference(
      { description: 'A long explanation nobody scans', reference: 'INV-4471' },
      'x',
    );
    expect(label).toBe('INV-4471');
  });

  it('accepts a number, because an invoice number is a reference', () => {
    expect(taskReference({ reference: 4471 }, 'x').label).toBe('4471');
  });

  it('reads the name whatever case the author used', () => {
    // A definition's author writes whatever they like. Being strict here would
    // fall through to the identifier for half of them.
    expect(taskReference({ Description: 'Expense claim' }, 'x').label).toBe('Expense claim');
    expect(taskReference({ REFERENCE: 'INV-1' }, 'x').label).toBe('INV-1');
  });
});

describe('variables are authored by somebody else', () => {
  /*
   * They come from a process definition a user wrote, so a value may be
   * anything. A list cell is not the place to find that out.
   */
  it('ignores values that are not text or a number', () => {
    expect(taskReference({ description: { nested: true } }, 'x').label).toBe('');
    expect(taskReference({ description: ['a', 'b'] }, 'x').label).toBe('');
    expect(taskReference({ description: true }, 'x').label).toBe('');
    expect(taskReference({ description: null }, 'x').label).toBe('');
  });

  it('ignores a value that is only whitespace', () => {
    expect(taskReference({ description: '   ' }, 'x').label).toBe('');
  });

  it('bounds a very long one, breaking on a word', () => {
    const long = 'Reimbursement for the offsite in Lisbon including flights and accommodation and meals';
    const { label } = taskReference({ description: long }, 'x');
    expect(label.length).toBeLessThanOrEqual(61);
    expect(label.endsWith('…')).toBe(true);
    expect(label).not.toContain('  ');
  });

  it('answers empty when there is nothing to say', () => {
    expect(taskReference(undefined, 'x').label).toBe('');
    expect(taskReference({}, 'x').label).toBe('');
    expect(taskReference({ amount: 10, approved: true }, 'x').label).toBe('');
  });
});

describe('the two parts together', () => {
  it('gives a caller both, so it can show the code when there is no label', () => {
    const ref = taskReference({}, '01a07764-6839-75b1-8ffe-ea705bb67a3b');
    expect(ref.label).toBe('');
    expect(ref.code).toBe('B67A3B');
  });
});
