/**
 * How a stored credential travels through the connector form without being
 * shown.
 *
 * The server no longer returns a connection's secrets. For any config key
 * whose name suggests one, it sends the sentinel below instead, and when it
 * receives that sentinel back on an update it keeps whatever it already has.
 * So the form's job is to leave the sentinel alone unless the person types a
 * replacement, and never to print it as though it were the value.
 */

/** What the server sends in place of a stored secret, and accepts back as "keep it". */
export const UNCHANGED_SECRET = '__unchanged__';

/** Placeholder shown where a stored secret would otherwise be. */
export const UNCHANGED_PLACEHOLDER = 'unchanged';

/** Config keys the server treats as secrets. Mirrors the server's list. */
const SENSITIVE_KEY = /secret|password|token|api_key|key/i;

/** Whether the server would mask this config key. */
export function isSensitiveKey(key: string): boolean {
  return SENSITIVE_KEY.test(key);
}

/** Whether a config value is the sentinel rather than something to show. */
export function isUnchanged(value: unknown): boolean {
  return value === UNCHANGED_SECRET;
}

/** What an input should display for a config value: never the sentinel. */
export function displayValue(value: unknown): string {
  if (value == null || isUnchanged(value)) return '';
  return String(value);
}

/**
 * The config keys still holding the sentinel.
 *
 * A test run sends the form's config as-is, and the execute endpoint has no
 * stored connection to fill the sentinel from — so a test with any of these
 * would send the literal word to the third party. The caller uses this to say
 * so rather than let that happen.
 */
export function unchangedKeys(config: Record<string, unknown> | undefined): string[] {
  if (!config) return [];
  return Object.entries(config)
    .filter(([, value]) => isUnchanged(value))
    .map(([key]) => key);
}
