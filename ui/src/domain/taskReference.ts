/**
 * How a piece of work identifies itself in a list.
 *
 * The task inbox showed the first eight characters of the process instance's
 * identifier. Those identifiers are time-ordered, so every row created in the
 * same period shares that prefix — three approvals in a row all read
 * `01a07764`. It identified nothing, took the widest column on the screen, and
 * the person reading it could not use it for anything.
 *
 * What somebody actually needs is *which* piece of work this is: "Expense of
 * GBP 1,750", not an identifier. The task carries the business variables the
 * process was started with, so the answer is usually already on screen —
 * it was just not being shown.
 *
 * The identifier still has a use, for quoting to support or matching against a
 * log, so a short code remains. It is taken from the *end* of the identifier,
 * where the characters differ, rather than the shared beginning.
 */

/** The variable names a process is most likely to describe itself with. */
const REFERENCE_KEYS = [
  'reference',
  'title',
  'subject',
  'description',
  'summary',
  'name',
  'label',
] as const;

/** Longer than this is a paragraph, not a reference, and breaks the layout. */
const MAX_LABEL = 60;

export interface TaskReference {
  /** What this piece of work is, in the words the process used. */
  label: string;
  /** A short code for quoting, or "" when there is no identifier. */
  code: string;
}

/**
 * Picks the business description of a task, and a short code for its instance.
 *
 * Variables come from a process definition somebody authored, so a value may be
 * anything at all: an object, an array, a number, a very long string. Only
 * strings and numbers are considered, and the result is bounded — a list cell
 * is not the place to discover that.
 */
export function taskReference(
  variables: Record<string, unknown> | undefined,
  instanceId: string | undefined,
): TaskReference {
  return { label: referenceLabel(variables), code: shortCode(instanceId) };
}

function referenceLabel(variables: Record<string, unknown> | undefined): string {
  if (!variables) return '';
  for (const key of REFERENCE_KEYS) {
    const value = pickText(variables, key);
    if (value !== '') return value;
  }
  return '';
}

/**
 * Reads one variable as display text, case-insensitively.
 *
 * Case-insensitive because a definition's author writes whatever they like —
 * `Description`, `DESCRIPTION`, `description` — and being strict here would
 * silently fall through to the identifier for half of them.
 */
function pickText(variables: Record<string, unknown>, key: string): string {
  for (const [name, value] of Object.entries(variables)) {
    if (name.toLowerCase() !== key) continue;
    if (typeof value === 'string') return truncate(value.trim());
    // A number is a legitimate reference — an invoice or ticket number — but a
    // boolean is not, and an object would render as "[object Object]".
    if (typeof value === 'number' && Number.isFinite(value)) return String(value);
    return '';
  }
  return '';
}

function truncate(text: string): string {
  if (text.length <= MAX_LABEL) return text;
  // Cut on a word where one is near the end, so the result reads as a phrase
  // rather than stopping mid-word.
  const cut = text.slice(0, MAX_LABEL);
  const lastSpace = cut.lastIndexOf(' ');
  return `${(lastSpace > MAX_LABEL - 15 ? cut.slice(0, lastSpace) : cut).trimEnd()}…`;
}

/**
 * The last characters of an identifier, which are the ones that differ.
 *
 * These are UUIDv7: the leading characters encode the time, so instances
 * created together share them. Taking the front — which is what the inbox used
 * to do — produces the same string for every row in a busy period.
 */
export function shortCode(instanceId: string | undefined): string {
  if (!instanceId) return '';
  const compact = instanceId.replace(/-/g, '');
  if (compact.length <= 6) return compact.toUpperCase();
  return compact.slice(-6).toUpperCase();
}
