import { describe, expect, it } from 'bun:test';
import { reconcileKeys, type KeyedId } from './stableKeys';

function minter() {
  let n = 0;
  return () => `k${(n += 1)}`;
}

describe('reconcileKeys', () => {
  it('keeps a key through a rename, so the card editing the id is not remounted', () => {
    const mint = minter();
    const first = reconcileKeys([], ['field_1', 'field_2'], mint);
    // Typing "a" into the first field's id.
    const second = reconcileKeys(first, ['field_1a', 'field_2'], mint);
    expect(second[0].key).toBe(first[0].key);
    expect(second[1].key).toBe(first[1].key);
  });

  it('follows a reorder by id', () => {
    const mint = minter();
    const first = reconcileKeys([], ['a', 'b', 'c'], mint);
    const moved = reconcileKeys(first, ['c', 'a', 'b'], mint);
    expect(moved.map((k) => k.key)).toEqual([first[2].key, first[0].key, first[1].key]);
  });

  it('mints a new key for an added field and forgets the key of a removed one', () => {
    const mint = minter();
    const first = reconcileKeys([], ['a', 'b'], mint);
    const added = reconcileKeys(first, ['a', 'b', 'new'], mint);
    expect(added.slice(0, 2)).toEqual(first);
    expect(first.map((k) => k.key)).not.toContain(added[2].key);

    const removed = reconcileKeys(added, ['a', 'new'], mint);
    expect(removed.map((k) => k.key)).toEqual([first[0].key, added[2].key]);
  });

  it('never hands the same key to two items, even with duplicate ids', () => {
    const previous: KeyedId[] = [{ id: 'x', key: 'old' }];
    const next = reconcileKeys(previous, ['x', 'x'], minter());
    expect(new Set(next.map((k) => k.key)).size).toBe(2);
  });
});
