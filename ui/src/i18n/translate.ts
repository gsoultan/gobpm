/**
 * Turning a message key and some values into a sentence.
 *
 * Written here rather than taken from a library for two reasons. The first-paint
 * budget is 336 kB gzipped and enforced by the build, and a full ICU runtime is
 * a meaningful share of it for a product that needs interpolation and plurals
 * and little else. The second is that this is decidable logic — given a
 * catalogue, a key and some values, there is one right answer — and the house
 * rule is that decidable logic lives in a module with tests rather than inside
 * a component.
 *
 * What it supports, deliberately and no more:
 *
 *   "Signed in as {name}"                      interpolation
 *   "{count, plural, one {# task} other {# tasks}}"   plurals, via Intl
 *
 * Plural categories come from `Intl.PluralRules`, so a locale with more than
 * two of them — Polish has four — is handled by the platform rather than by a
 * rule this file would get wrong.
 */

/** A flat catalogue. Flat, so a missing key is a lookup rather than a walk. */
export type Catalogue = Record<string, string>;

/** Values interpolated into a message. */
export type Values = Record<string, string | number>;

const PLACEHOLDER = /\{(\w+)\}/g;
const PLURAL = /\{(\w+),\s*plural,\s*([\s\S]*?)\}\s*$/;

/**
 * Formats one message.
 *
 * A key with no entry returns the key itself. That is deliberate: a blank space
 * where a word should be is a bug nobody can see, and an English fallback would
 * hide a missing translation from the person who could fix it. The key is ugly
 * on screen, which is the point.
 */
export function format(catalogue: Catalogue, key: string, values: Values = {}): string {
  const message = catalogue[key];
  if (message === undefined) return key;
  const { text, count } = applyPlural(message, values);
  return interpolate(text, values, count);
}

/**
 * Resolves a plural form.
 *
 * The whole message is either a plural or it is not; nesting one inside a
 * sentence is where message formats stop being readable to translators, and
 * every case here is a whole sentence anyway.
 */
function applyPlural(message: string, values: Values): { text: string; count?: number } {
  const match = PLURAL.exec(message.trim());
  if (!match) return { text: message };

  const [, name, body] = match;
  const count = Number(values[name]);
  if (!Number.isFinite(count)) return { text: message };

  const forms = parseForms(body);
  // An exact match wins over a category: "=0" is how a language says "no
  // tasks" rather than "0 tasks", which is a different sentence in most of them.
  const exact = forms[`=${count}`];
  if (exact !== undefined) return { text: exact, count };

  const category = pluralCategory(count);
  return { text: forms[category] ?? forms.other ?? message, count };
}

/** `one {# task} other {# tasks}` → `{ one: '# task', other: '# tasks' }`. */
function parseForms(body: string): Record<string, string> {
  const forms: Record<string, string> = {};
  let index = 0;
  while (index < body.length) {
    const open = body.indexOf('{', index);
    if (open === -1) break;
    const name = body.slice(index, open).trim();

    // Braces nest, so scanning to the first '}' would cut a form short.
    let depth = 1;
    let cursor = open + 1;
    while (cursor < body.length && depth > 0) {
      if (body[cursor] === '{') depth += 1;
      else if (body[cursor] === '}') depth -= 1;
      cursor += 1;
    }
    if (name) forms[name] = body.slice(open + 1, cursor - 1);
    index = cursor;
  }
  return forms;
}

let rules: Intl.PluralRules | undefined;
let rulesLocale = '';

function pluralCategory(count: number): string {
  const locale = currentLocaleTag;
  if (!rules || rulesLocale !== locale) {
    try {
      rules = new Intl.PluralRules(locale);
    } catch {
      rules = new Intl.PluralRules('en');
    }
    rulesLocale = locale;
  }
  return rules.select(count);
}

/**
 * The locale plural selection uses.
 *
 * Set by the provider. Held here rather than threaded through every call so
 * that `format` stays a two-argument function at every call site, which is what
 * keeps the components readable.
 */
let currentLocaleTag = 'en';

export function setFormattingLocale(locale: string): void {
  currentLocaleTag = locale;
}

function interpolate(message: string, values: Values, count?: number): string {
  // `#` stands for the count, and *only* inside a plural form — which is why
  // the count is passed rather than guessed. Replacing it everywhere would
  // rewrite an ordinary "Ticket #4471" into something else entirely.
  const withCount = count === undefined ? message : message.replace(/#/g, String(count));
  return withCount
    .replace(PLACEHOLDER, (whole, name: string) => {
      const value = values[name];
      // An unresolved placeholder is left as it was written, so it is obvious
      // in a screenshot rather than rendering as "undefined".
      return value === undefined ? whole : String(value);
    });
}
