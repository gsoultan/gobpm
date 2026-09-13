/**
 * Whether a row matches what somebody typed into a search box.
 *
 * Every list page had its own copy of this — or, on three of them, a search
 * box with no handler at all. One definition means one answer to "is the
 * match case-sensitive" (no) and "does an empty box hide anything" (no).
 */
export function matchesQuery(query: string, ...fields: Array<string | undefined | null>): boolean {
  const needle = query.trim().toLowerCase();
  if (!needle) return true;
  return fields.some((field) => field != null && field.toLowerCase().includes(needle));
}
