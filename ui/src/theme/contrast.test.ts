import { describe, expect, it } from 'bun:test';

import {
  AA_NON_TEXT,
  AA_NORMAL_TEXT,
  contrastRatio,
  meets,
  parseHex,
  relativeLuminance,
} from './contrast';
import { DIMMED_DARK, DIMMED_LIGHT, SURFACE_DARK, SURFACE_LIGHT, SURFACE_PAGE } from './palette';

describe('the arithmetic', () => {
  it('matches the reference values in the specification', () => {
    expect(relativeLuminance('#ffffff')).toBeCloseTo(1, 5);
    expect(relativeLuminance('#000000')).toBeCloseTo(0, 5);
    // Black on white is the maximum a screen can do.
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 2);
    // A colour against itself has no contrast at all.
    expect(contrastRatio('#777777', '#777777')).toBeCloseTo(1, 5);
  });

  it('is symmetric, because contrast has no direction', () => {
    expect(contrastRatio('#495057', '#ffffff')).toBeCloseTo(contrastRatio('#ffffff', '#495057'), 10);
  });

  it('reads both hex forms and refuses anything else', () => {
    expect(parseHex('#fff')).toEqual([255, 255, 255]);
    expect(parseHex('1c7ed6')).toEqual([28, 126, 214]);
    expect(() => parseHex('rebeccapurple')).toThrow();
    expect(() => parseHex('#12345')).toThrow();
  });
});

/**
 * The regression this file exists for.
 *
 * Secondary text shipped at 3.15:1 against the page background where AA asks
 * for 4.5:1 — Mantine's default dimmed grey, which is fine on white and not on
 * the slightly grey surface the app actually uses. Contrast is the one design
 * property that cannot be checked by looking: it looked fine to whoever chose
 * it and was unreadable to somebody with low vision or a screen in sunlight.
 */
describe('secondary text is readable', () => {
  it('passes AA on every surface it is used on, in light mode', () => {
    for (const surface of [SURFACE_LIGHT, SURFACE_PAGE]) {
      const ratio = contrastRatio(DIMMED_LIGHT, surface);
      expect(
        meets(DIMMED_LIGHT, surface, AA_NORMAL_TEXT),
        `${DIMMED_LIGHT} on ${surface} is ${ratio.toFixed(2)}:1, AA needs ${AA_NORMAL_TEXT}`,
      ).toBe(true);
    }
  });

  it('passes AA in dark mode', () => {
    const ratio = contrastRatio(DIMMED_DARK, SURFACE_DARK);
    expect(
      meets(DIMMED_DARK, SURFACE_DARK, AA_NORMAL_TEXT),
      `${DIMMED_DARK} on ${SURFACE_DARK} is ${ratio.toFixed(2)}:1`,
    ).toBe(true);
  });

  /*
   * Guards against the tempting near-miss.
   *
   * Mantine's own next-darker grey measures 4.45:1 on the page background: it
   * reads as "basically 4.5" and fails. The threshold is not a suggestion, and
   * the surface the app uses is not white.
   */
  it('rejects the greys that only pass against pure white', () => {
    expect(meets('#868e96', SURFACE_PAGE, AA_NORMAL_TEXT)).toBe(false);
    expect(meets('#6c757d', SURFACE_PAGE, AA_NORMAL_TEXT)).toBe(false);
    expect(meets('#6c757d', SURFACE_LIGHT, AA_NORMAL_TEXT)).toBe(true);
  });

  it('keeps secondary text visibly secondary', () => {
    // Passing by going black would meet the letter of the rule and lose the
    // distinction the colour exists to make.
    expect(contrastRatio(DIMMED_LIGHT, SURFACE_PAGE)).toBeLessThan(
      contrastRatio('#000000', SURFACE_PAGE),
    );
  });
});

describe('the brand colour', () => {
  it('is usable as a non-text element on light surfaces', () => {
    // Borders, icons and focus rings drawn in the brand colour are interface
    // components, which AA holds to 3:1.
    for (const surface of [SURFACE_LIGHT, SURFACE_PAGE]) {
      expect(meets('#1c7ed6', surface, AA_NON_TEXT)).toBe(true);
    }
  });
});
