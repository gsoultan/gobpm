/**
 * Keys for a list whose items are identified by something the user can edit.
 *
 * The form builder keyed each field card on `field.id`, and the card contains
 * the input that edits `field.id`. Every keystroke changed the key, React
 * unmounted the card and mounted a new one, and the input lost focus after a
 * single character. The key has to be something the user cannot type into.
 *
 * Keys are minted once and follow an item through renames: an item at the
 * same position whose old id has disappeared from the list is the same item
 * with a new name, not a new item.
 */
export interface KeyedId {
  id: string;
  key: string;
}

export function reconcileKeys(previous: KeyedId[], currentIds: string[], mint: () => string): KeyedId[] {
  const previousById = new Map(previous.map((entry) => [entry.id, entry]));
  const currentIdSet = new Set(currentIds);
  const usedKeys = new Set<string>();

  const claim = (entry: KeyedId | undefined): string | null => {
    if (!entry || usedKeys.has(entry.key)) return null;
    usedKeys.add(entry.key);
    return entry.key;
  };

  return currentIds.map((id, index) => {
    const sameId = claim(previousById.get(id));
    if (sameId) return { id, key: sameId };

    const atPosition = previous[index];
    const renamedInPlace = atPosition && !currentIdSet.has(atPosition.id) ? claim(atPosition) : null;
    if (renamedInPlace) return { id, key: renamedInPlace };

    const fresh = mint();
    usedKeys.add(fresh);
    return { id, key: fresh };
  });
}
