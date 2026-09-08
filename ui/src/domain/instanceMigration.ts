import type { ApiMigrationPlan, ApiNodeMove } from '../services/types';

/**
 * Reading a migration plan.
 *
 * Moving running instances rewrites work that is already somebody's — a
 * purchase order halfway through approval, a leave request sitting in an
 * inbox. The decidable parts of presenting that live here rather than in the
 * modal, because "would this strand anybody's task" is a question with an
 * answer, not a rendering detail.
 */

/** How much work sits on one node, across all three kinds. */
export function workOn(move: ApiNodeMove): number {
  return move.tokens + move.tasks + move.jobs;
}

/** The nodes that would actually move, in plan order. */
export function movedNodes(plan: ApiMigrationPlan): ApiNodeMove[] {
  return (plan.moves ?? []).filter((move) => move.from !== move.to);
}

/**
 * The nodes carried across unchanged because the new version still has them.
 *
 * Shown, not hidden: this is how somebody confirms they did not need a mapping
 * for a node rather than forgot one.
 */
export function carriedNodes(plan: ApiMigrationPlan): ApiNodeMove[] {
  return (plan.moves ?? []).filter((move) => move.from === move.to);
}

/** Whether applying this plan would be accepted. */
export function isApplicable(plan: ApiMigrationPlan | null): boolean {
  return plan !== null && (plan.refusals?.length ?? 0) === 0;
}

/** Total tasks that would move — the number that means "people affected". */
export function tasksAffected(plan: ApiMigrationPlan): number {
  return (plan.moves ?? []).reduce((total, move) => total + move.tasks, 0);
}

/**
 * One sentence for the top of the dialog.
 *
 * Written to be readable when the answer is "nothing": a plan over zero
 * instances is the common case once a version has drained, and telling somebody
 * "0 instances would move" is clearer than an empty table.
 */
export function planSummary(plan: ApiMigrationPlan): string {
  if (plan.instances === 0) {
    return `Nothing is running on version ${plan.source_version}. There is nothing to move.`;
  }
  const instances = plan.instances === 1 ? '1 instance' : `${plan.instances} instances`;
  const tasks = tasksAffected(plan);
  if (tasks === 0) {
    return `${instances} would move from version ${plan.source_version} to version ${plan.target_version}.`;
  }
  const inboxes = tasks === 1 ? '1 task' : `${tasks} tasks`;
  return `${instances} would move from version ${plan.source_version} to version ${plan.target_version}, including ${inboxes} already in somebody's inbox.`;
}

/**
 * Turns the mapping rows a person edits into what the API takes.
 *
 * Blank targets are dropped rather than sent as empty strings: leaving a row
 * empty means "I have not decided", and sending it would ask the server to move
 * work onto a node called "".
 */
export function toNodeMapping(rows: readonly { from: string; to: string }[]): Record<string, string> {
  const mapping: Record<string, string> = {};
  for (const row of rows) {
    const from = row.from.trim();
    const to = row.to.trim();
    if (from === '' || to === '' || from === to) continue;
    mapping[from] = to;
  }
  return mapping;
}
