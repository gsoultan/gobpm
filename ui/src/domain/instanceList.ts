/**
 * What a row in the instance list says about an instance.
 *
 * A listed instance carries an id, a status, its active nodes and a shell of
 * its definition. None of that is what a person identifies it by, so each
 * function here turns one of those fields into something they can read.
 */

/** As much of an instance as the list needs. */
export interface ListedInstance {
  id: string;
  status?: string;
  definition?: { id?: string; key?: string; name?: string };
}

/** As much of a definition as resolving a name needs. */
export interface NamedDefinition {
  id: string;
  key?: string;
  name?: string;
}

/** Shown when neither the instance nor the directory can name the process. */
export const UNNAMED_PROCESS = 'Process';

/** The name of the process an instance belongs to, resolved through its id. */
export function definitionName(instance: ListedInstance, definitions: NamedDefinition[]): string {
  const fromInstance = instance.definition?.name || instance.definition?.key;
  if (fromInstance) return fromInstance;
  const match = definitions.find((d) => d.id === instance.definition?.id);
  return match?.name || match?.key || UNNAMED_PROCESS;
}

/**
 * Turns a node identifier into something readable.
 *
 * Node IDs are authored in the designer and usually carry the step's name in
 * them — "Activity_ApproveExpense", "approve-expense", "Task_1". Splitting the
 * generated prefix and the separators recovers a usable label without needing
 * the whole definition loaded just to render a row.
 */
export function humanizeNodeId(nodeId: string): string {
  const withoutPrefix = nodeId.replace(/^(Activity|Task|Event|Gateway|Flow|Node)[_-]/i, '');
  const spaced = withoutPrefix
    .replace(/[_-]+/g, ' ')
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .trim();
  if (!spaced || /^\d+$/.test(spaced)) return nodeId;
  return spaced.charAt(0).toUpperCase() + spaced.slice(1);
}

/** The version nibble a time-ordered UUID carries. */
const UUID_V7 = '7';
/** How many hex digits of a v7 id are the millisecond timestamp. */
const V7_TIMESTAMP_HEX_DIGITS = 12;
/** How many trailing hex digits make a reference short enough to read out. */
const REFERENCE_LENGTH = 6;

/**
 * When an instance started, read out of its id.
 *
 * Every primary key here is a UUIDv7, whose first 48 bits are the creation
 * time in milliseconds. The list does not carry a started-at field, and the
 * eight-character prefix it used to print was that timestamp's most
 * significant digits — identical on every row created the same week.
 *
 * Returns null for anything that is not a v7 id rather than inventing a date.
 */
export function startedAtFromId(id: string): Date | null {
  const hex = id.replace(/-/g, '');
  if (hex.length !== 32 || hex[V7_TIMESTAMP_HEX_DIGITS] !== UUID_V7) return null;
  const millis = Number.parseInt(hex.slice(0, V7_TIMESTAMP_HEX_DIGITS), 16);
  if (Number.isNaN(millis)) return null;
  return new Date(millis);
}

/**
 * A short reference that differs between rows.
 *
 * The tail of a v7 id is random, so its last digits distinguish two instances
 * started in the same millisecond where the head cannot. Short enough to read
 * over the phone to whoever is looking at the same list.
 */
export function instanceReference(id: string): string {
  const hex = id.replace(/-/g, '');
  return `#${hex.slice(-REFERENCE_LENGTH).toUpperCase()}`;
}

/** The distinct statuses on this page, in the order they first appear. */
export function statusesOnPage(instances: Array<{ status?: string }>): string[] {
  const seen = new Set<string>();
  for (const instance of instances) {
    const status = instance.status?.toLowerCase();
    if (status) seen.add(status);
  }
  return [...seen];
}

/** The rows matching a status; every row when no status is chosen. */
export function withStatus<T extends { status?: string }>(instances: T[], status: string | null): T[] {
  if (!status) return instances;
  return instances.filter((instance) => instance.status?.toLowerCase() === status);
}
