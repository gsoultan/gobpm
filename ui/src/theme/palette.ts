/**
 * The colours whose contrast is asserted.
 *
 * They live here, apart from the Mantine theme, so the check in
 * `contrast.test.ts` reads the same values the stylesheet applies rather than a
 * copy that can drift. Changing one of these changes what the test measures.
 */

/** Cards, modals, tables: the surface most text sits on in light mode. */
export const SURFACE_LIGHT = '#ffffff';

/**
 * The application background in light mode.
 *
 * The important one, and the one that was missed. It is slightly grey, so a
 * foreground chosen against pure white loses roughly 0.2 of its ratio here —
 * enough to take Mantine's default dimmed grey from "nearly passing" to
 * failing, on the surface most of the page actually uses.
 */
export const SURFACE_PAGE = '#f8f9fa';

/** The application background in dark mode (Mantine dark-7). */
export const SURFACE_DARK = '#1a1b1e';

/**
 * Secondary text in light mode.
 *
 * Mantine's default is `#868e96`, which measures 3.15:1 on the page background
 * where AA asks for 4.5. This is dark enough to pass on both light surfaces
 * (5.78:1 on the page, 6.09:1 on a card) and still clearly quieter than body
 * text.
 */
export const DIMMED_LIGHT = '#5c636a';

/** Secondary text in dark mode. Mantine's default already passes, at 7.16:1. */
export const DIMMED_DARK = '#a6a7ab';
