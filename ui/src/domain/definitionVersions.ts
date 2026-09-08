/**
 * Which version of each process the list shows.
 *
 * Every deployment is its own definition row sharing a key with the ones
 * before it. The list shows one row per key — the newest — and keeps the
 * rest for the version history.
 */

/** As much of a definition as grouping by version needs. */
export interface VersionedDefinition {
  key: string;
  version: number;
}

/** Every version of each key, newest first. */
export function versionsByKey<T extends VersionedDefinition>(definitions: T[]): Record<string, T[]> {
  const groups: Record<string, T[]> = {};
  for (const definition of definitions) {
    (groups[definition.key] ??= []).push(definition);
  }
  for (const versions of Object.values(groups)) {
    versions.sort((a, b) => b.version - a.version);
  }
  return groups;
}

/** The newest version of each key, in the order the keys first appear. */
export function latestVersions<T extends VersionedDefinition>(definitions: T[]): T[] {
  return Object.values(versionsByKey(definitions)).map((versions) => versions[0]);
}
