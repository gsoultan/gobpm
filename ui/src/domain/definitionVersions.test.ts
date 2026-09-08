import { describe, expect, it } from 'bun:test';

import { latestVersions, versionsByKey } from './definitionVersions';

const rows = [
  { id: 'a1', key: 'expense', version: 1 },
  { id: 'b2', key: 'onboarding', version: 2 },
  { id: 'a3', key: 'expense', version: 3 },
  { id: 'a2', key: 'expense', version: 2 },
];

describe('versionsByKey', () => {
  it('groups every version under its key, newest first', () => {
    const groups = versionsByKey(rows);
    expect(groups.expense.map((d) => d.version)).toEqual([3, 2, 1]);
    expect(groups.onboarding.map((d) => d.version)).toEqual([2]);
  });
});

describe('latestVersions', () => {
  it('keeps one row per key: the newest', () => {
    expect(latestVersions(rows).map((d) => d.id)).toEqual(['a3', 'b2']);
  });

  it('is empty for an empty page', () => {
    expect(latestVersions([])).toEqual([]);
  });
});
