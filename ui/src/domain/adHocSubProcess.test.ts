import { describe, expect, it } from 'bun:test';

import { adHocBlocks, adHocToggle, adHocWarnings } from './adHocSubProcess';

describe('adHocToggle', () => {
  it('clears the completion condition when ad-hoc is turned off', () => {
    // Left behind, it is a field the panel no longer shows and the server still
    // stores — invisible configuration that survives a deploy.
    expect(adHocToggle(false)).toEqual({ isAdHoc: false, completionCondition: '' });
  });

  it('keeps whatever condition is already there when ad-hoc is turned on', () => {
    expect(adHocToggle(true)).toEqual({ isAdHoc: true });
  });
});

describe('adHocWarnings', () => {
  it('says nothing about an ordinary sub-process', () => {
    expect(adHocWarnings({ isAdHoc: false }, 0)).toEqual([]);
  });

  it('refuses an ad-hoc sub-process with nothing in it', () => {
    const warnings = adHocWarnings({ isAdHoc: true, completionCondition: 'done' }, 0);
    expect(warnings).toHaveLength(1);
    expect(warnings[0].field).toBe('steps');
    expect(warnings[0].severity).toBe('error');
  });

  it('warns that no condition means the process does not wait at all', () => {
    const warnings = adHocWarnings({ isAdHoc: true }, 3);
    expect(warnings.map((w) => w.field)).toEqual(['completionCondition']);
    expect(warnings[0].severity).toBe('warning');
  });

  it('treats blank space as no condition', () => {
    expect(adHocWarnings({ isAdHoc: true, completionCondition: '   ' }, 3)).toHaveLength(1);
  });

  it('refuses a sub-process that is both event-triggered and ad-hoc', () => {
    const warnings = adHocWarnings(
      { isAdHoc: true, isEventSubProcess: true, completionCondition: 'done' },
      2,
    );
    expect(warnings.map((w) => w.field)).toEqual(['isEventSubProcess']);
    expect(warnings[0].severity).toBe('error');
  });

  it('reports everything wrong at once rather than one thing at a time', () => {
    const warnings = adHocWarnings({ isAdHoc: true, isEventSubProcess: true }, 0);
    expect(warnings.map((w) => w.field)).toEqual(['steps', 'completionCondition', 'isEventSubProcess']);
  });
});

describe('adHocBlocks', () => {
  it('is false when the only complaint is a caution', () => {
    expect(adHocBlocks({ isAdHoc: true }, 2)).toBe(false);
  });

  it('is true when the sub-process cannot work as configured', () => {
    expect(adHocBlocks({ isAdHoc: true, completionCondition: 'done' }, 0)).toBe(true);
  });
});
