/**
 * Which organization and project the app is currently working in.
 *
 * Signing in used to leave both unset, so the first thing anybody saw after
 * entering their password was "Choose a project to continue" — on every screen,
 * every session, even when they belonged to exactly one. A question the app
 * could answer itself is not a choice; it is a step.
 *
 * The rules, in order:
 *
 *   1. Keep what is already selected, if it is still available. Somebody who
 *      picked a project must not be moved off it on the next render.
 *   2. Otherwise take the first available. That covers a fresh sign-in, and
 *      also the case where the stored selection has gone — a project deleted,
 *      or access to it withdrawn — where the alternative is a session pinned to
 *      something the user can no longer see.
 *   3. Only when there is genuinely nothing available is the answer "none",
 *      and the interface then asks them to create one.
 *
 * Kept here rather than in the layout so it can be tested without a browser:
 * the "still loading" case below is exactly the sort of thing that is obvious
 * in a unit test and invisible in a render.
 */

/** The little a selectable thing has to have. */
export interface Selectable {
  id: string;
}

/**
 * The selection to apply, given what is available and what is stored.
 *
 * `available` must be the *loaded* list. An empty array is read as "there are
 * none", so passing an empty one while the request is still in flight would
 * clear a perfectly good selection and flash the empty state — see
 * `hasLoaded` at the call site.
 */
export function resolveSelection(available: readonly Selectable[], current: string | null): string | null {
  if (current !== null && available.some((item) => item.id === current)) {
    return current;
  }
  return available.length > 0 ? available[0].id : null;
}

/**
 * Whether the store needs updating, so the caller can skip a write that would
 * set the same value.
 *
 * A no-op write to the store is not free here: it re-renders every component
 * subscribed to it, and this runs from an effect that depends on the value it
 * would be setting.
 */
export function selectionNeedsUpdate(
  available: readonly Selectable[],
  current: string | null,
): boolean {
  return resolveSelection(available, current) !== current;
}
