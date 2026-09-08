/**
 * WCAG contrast, computed rather than eyeballed.
 *
 * The interface shipped with secondary text at 3.17:1 against the page
 * background, where AA asks for 4.5:1. Nobody chose that — it is Mantine's
 * default dimmed grey, which is fine on white and not on the slightly grey
 * surface the app actually uses. It went unnoticed because contrast is the one
 * design property you cannot check by looking: it looks fine to whoever picked
 * it, and is unreadable to somebody with low vision or a screen in sunlight.
 *
 * These are the numbers, so a colour change that breaks them fails a test
 * instead of shipping.
 *
 * Reference: WCAG 2.1 §1.4.3, using the relative luminance formula from
 * https://www.w3.org/TR/WCAG21/#dfn-relative-luminance
 */

/** AA for body text. */
export const AA_NORMAL_TEXT = 4.5;
/** AA for text at 18.66px bold or 24px regular and above. */
export const AA_LARGE_TEXT = 3;
/** AA for interface components and meaningful graphics. */
export const AA_NON_TEXT = 3;

/** Parses `#rgb` or `#rrggbb` into 0-255 channels. */
export function parseHex(hex: string): [number, number, number] {
  const text = hex.trim().replace(/^#/, '');
  const full =
    text.length === 3
      ? text
          .split('')
          .map((c) => c + c)
          .join('')
      : text;
  if (!/^[0-9a-fA-F]{6}$/.test(full)) {
    throw new Error(`not a hex colour: ${hex}`);
  }
  return [
    parseInt(full.slice(0, 2), 16),
    parseInt(full.slice(2, 4), 16),
    parseInt(full.slice(4, 6), 16),
  ];
}

/** The relative luminance of a colour, 0 for black and 1 for white. */
export function relativeLuminance(hex: string): number {
  const [r, g, b] = parseHex(hex).map((channel) => {
    const c = channel / 255;
    // The linearisation the specification defines; the threshold and exponent
    // are not adjustable.
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  }) as [number, number, number];
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** The contrast ratio between two colours, from 1 (identical) to 21. */
export function contrastRatio(foreground: string, background: string): number {
  const a = relativeLuminance(foreground);
  const b = relativeLuminance(background);
  const lighter = Math.max(a, b);
  const darker = Math.min(a, b);
  return (lighter + 0.05) / (darker + 0.05);
}

/** Whether a pair meets a threshold, rounded the way a checker reports it. */
export function meets(foreground: string, background: string, threshold: number): boolean {
  // Rounded to two places first: a checker reports 4.5, and failing a pair that
  // reports as passing would be a test nobody can act on.
  return Math.round(contrastRatio(foreground, background) * 100) / 100 >= threshold;
}
