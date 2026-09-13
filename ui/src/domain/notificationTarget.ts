/**
 * Where a notification takes the person who opens it.
 *
 * A notification about a task points at the inbox, because that is where the
 * task is acted on. One that only names a process instance opens that
 * instance in the designer, which is where an instance is looked at. One that
 * names nothing goes nowhere — it is read, not followed.
 *
 * The server writes `link` as "/tasks?id=<node>", and nothing reads the query
 * string: the inbox lists the user's tasks regardless. The path is what is
 * honoured, so the router still gets a route it knows rather than a raw
 * string.
 */

/** The fields of a notification that decide where it leads. */
export interface NotificationLead {
  link?: string;
  instance_id?: string;
}

export type NotificationTarget =
  | { to: "/tasks" }
  | { to: "/designer"; search: { instanceId: string } };

const TASK_INBOX_PATH = "/tasks";

export function notificationTarget(notification: NotificationLead): NotificationTarget | null {
  const path = notification.link?.split("?")[0] ?? "";
  if (path === TASK_INBOX_PATH) {
    return { to: TASK_INBOX_PATH };
  }
  if (notification.instance_id) {
    return { to: "/designer", search: { instanceId: notification.instance_id } };
  }
  return null;
}
