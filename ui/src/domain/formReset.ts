/**
 * When a task form should throw away what the user has typed.
 *
 * The form used to reset whenever the `fields` or `variables` prop changed
 * identity. Both are derived from the task — the fields by parsing a JSON
 * string, the variables by `task.variables || {}` — so every background
 * refetch produced fresh objects with the same content, and the inbox wiped a
 * half-filled form every thirty seconds.
 *
 * The decision here is by content: a reset is warranted only when the
 * definition or the starting values actually differ.
 */

/** A fingerprint of the inputs a form was built from. */
export function formIdentity(fields: unknown, variables: unknown): string {
  return `${safeStringify(fields)}|${safeStringify(variables)}`;
}

/** True when the form must be rebuilt because its inputs changed in substance. */
export function shouldResetForm(previousIdentity: string, nextIdentity: string): boolean {
  return previousIdentity !== nextIdentity;
}

function safeStringify(value: unknown): string {
  try {
    return JSON.stringify(value) ?? '';
  } catch {
    // A circular structure cannot be fingerprinted; treating it as always
    // different is the pre-existing behaviour, and safer than never resetting.
    return String(Math.random());
  }
}
