/**
 * Reading a task's `form_definition` string into fields.
 *
 * The inbox parsed it inline, on every render, and turned a parse failure into
 * an empty form — which renders as "No inputs required" with a Complete
 * button. A form nobody could read was completing tasks with no input. A
 * failed read must say so.
 */
export interface ParsedFormDefinition<TField> {
  fields: TField[];
  /** Why the definition could not be read, in words; null when it could. */
  error: string | null;
}

export function parseFormDefinition<TField>(definition: string | undefined | null): ParsedFormDefinition<TField> {
  if (!definition || definition.trim() === '') {
    return { fields: [], error: null };
  }
  try {
    const parsed: unknown = JSON.parse(definition);
    if (Array.isArray(parsed)) {
      return { fields: parsed as TField[], error: null };
    }
    if (parsed && typeof parsed === 'object' && Array.isArray((parsed as { fields?: unknown }).fields)) {
      return { fields: (parsed as { fields: TField[] }).fields, error: null };
    }
    return { fields: [], error: "This task's form is not a list of fields." };
  } catch {
    return { fields: [], error: "This task's form could not be read. Ask whoever built the process to check it." };
  }
}
