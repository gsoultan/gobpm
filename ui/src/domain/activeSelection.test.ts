import { describe, expect, it } from 'bun:test';

import { resolveSelection, selectionNeedsUpdate } from './activeSelection';

const projects = [{ id: 'p1' }, { id: 'p2' }, { id: 'p3' }];

describe('picking one automatically', () => {
  /*
   * The reason this exists. Signing in left nothing selected, so the first
   * thing anybody saw after entering their password was "Choose a project to
   * continue" — on every screen, every session, even with exactly one project.
   */
  it('takes the first when nothing is selected', () => {
    expect(resolveSelection(projects, null)).toBe('p1');
  });

  it('takes the only one when there is only one', () => {
    expect(resolveSelection([{ id: 'solo' }], null)).toBe('solo');
  });

  it('answers none when there is genuinely nothing to pick', () => {
    // The interface then asks them to create one, which is the honest state
    // for a brand new account.
    expect(resolveSelection([], null)).toBeNull();
  });
});

describe('not fighting the person using it', () => {
  it('keeps a selection that is still available', () => {
    // Snapping back to the first on the next render would make choosing
    // anything else impossible.
    expect(resolveSelection(projects, 'p3')).toBe('p3');
  });

  it('is stable across repeated calls', () => {
    let current = resolveSelection(projects, null);
    for (let i = 0; i < 5; i += 1) {
      current = resolveSelection(projects, current);
    }
    expect(current).toBe('p1');
  });
});

describe('when the stored selection has gone', () => {
  /*
   * A project can be deleted, or access to it withdrawn. Leaving the session
   * pinned to it means every list answers empty and nothing says why.
   */
  it('falls back to the first available', () => {
    expect(resolveSelection(projects, 'deleted-project')).toBe('p1');
  });

  it('falls back to none when nothing is left', () => {
    expect(resolveSelection([], 'deleted-project')).toBeNull();
  });
});

describe('knowing whether to write to the store', () => {
  // A no-op write re-renders everything subscribed to the store, from an effect
  // that depends on the value it would be setting.
  it('reports no update needed when the selection already holds', () => {
    expect(selectionNeedsUpdate(projects, 'p2')).toBe(false);
    expect(selectionNeedsUpdate([], null)).toBe(false);
  });

  it('reports an update needed when there is nothing selected', () => {
    expect(selectionNeedsUpdate(projects, null)).toBe(true);
  });

  it('reports an update needed when the selection is gone', () => {
    expect(selectionNeedsUpdate(projects, 'gone')).toBe(true);
  });
});
