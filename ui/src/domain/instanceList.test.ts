import { describe, expect, it } from 'bun:test';

import {
  definitionName,
  humanizeNodeId,
  instanceReference,
  startedAtFromId,
  statusesOnPage,
  withStatus,
} from './instanceList';

describe('definitionName', () => {
  it('prefers what the instance itself carries', () => {
    expect(definitionName({ id: 'i', definition: { id: 'd', name: 'Expense claim' } }, [])).toBe('Expense claim');
    expect(definitionName({ id: 'i', definition: { id: 'd', key: 'expense' } }, [])).toBe('expense');
  });

  it('resolves through the directory when the instance only has an id', () => {
    const directory = [{ id: 'd', key: 'expense', name: 'Expense claim' }];
    expect(definitionName({ id: 'i', definition: { id: 'd' } }, directory)).toBe('Expense claim');
  });

  it('falls back to a generic word rather than an id', () => {
    expect(definitionName({ id: 'i', definition: { id: 'missing' } }, [])).toBe('Process');
  });
});

describe('humanizeNodeId', () => {
  it('drops the designer prefix and splits the words', () => {
    expect(humanizeNodeId('Activity_ApproveExpense')).toBe('Approve Expense');
    expect(humanizeNodeId('approve-expense')).toBe('Approve expense');
  });

  it('keeps a generated id that has no words in it', () => {
    expect(humanizeNodeId('Task_1')).toBe('Task_1');
  });
});

describe('startedAtFromId', () => {
  // 0x018f_0000_0000 ms since the epoch = 2024-04-21T09:32:31.104Z: the first 48 bits of a v7 id.
  const v7 = '018f0000-0000-7abc-8def-0123456789ab';

  it('reads the creation time out of a v7 id', () => {
    expect(startedAtFromId(v7)?.toISOString()).toBe('2024-04-21T09:32:31.104Z');
  });

  it('refuses to invent a time for an id that is not v7', () => {
    // Same digits, version nibble 4: a random UUID carries no timestamp.
    expect(startedAtFromId('018f0000-0000-4abc-8def-0123456789ab')).toBeNull();
    expect(startedAtFromId('not-an-id')).toBeNull();
  });
});

describe('instanceReference', () => {
  it('takes the random tail, which differs between rows the head does not', () => {
    // The bug: every UUIDv7 created the same week shares its first 8 chars,
    // so the old `id.slice(0, 8)` subtitle read identically on every row.
    const a = '018f0000-0000-7abc-8def-0123456789ab';
    const b = '018f0000-0000-7abc-8def-0123456789cd';
    expect(a.slice(0, 8)).toBe(b.slice(0, 8));
    expect(instanceReference(a)).toBe('#6789AB');
    expect(instanceReference(b)).toBe('#6789CD');
  });
});

describe('status filter', () => {
  const rows = [
    { id: '1', status: 'active' },
    { id: '2', status: 'failed' },
    { id: '3', status: 'ACTIVE' },
    { id: '4' },
  ];

  it('lists each status once, in the order it first appears', () => {
    expect(statusesOnPage(rows)).toEqual(['active', 'failed']);
  });

  it('keeps the rows with the chosen status, case-insensitively', () => {
    expect(withStatus(rows, 'active').map((r) => r.id)).toEqual(['1', '3']);
  });

  it('keeps every row when nothing is chosen', () => {
    expect(withStatus(rows, null)).toBe(rows);
  });
});
