/**
 * Turning an identifier into words.
 *
 * Keys such as `approval_level`, `waitingForInput` or `decision-key` leak out
 * of the engine into badges and sentences. Nobody outside engineering should
 * have to read them; this is the one place they are turned into "Approval
 * level", "Waiting for input" and "Decision key".
 */
export function humanizeIdentifier(identifier: string): string {
  const spaced = identifier
    .trim()
    .replace(/[_-]+/g, ' ')
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/\s+/g, ' ')
    .toLowerCase();
  if (spaced === '') {
    return '';
  }
  return spaced[0].toUpperCase() + spaced.slice(1);
}
