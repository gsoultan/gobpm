import { describe, expect, test } from "bun:test";
import { notificationTarget } from "./notificationTarget";

describe("notificationTarget", () => {
  test("a task notification leads to the inbox, whatever the query string says", () => {
    expect(notificationTarget({ link: "/tasks?id=approve", instance_id: "i-1" })).toEqual({ to: "/tasks" });
  });

  test("a notification that only names an instance opens that instance", () => {
    expect(notificationTarget({ instance_id: "i-1" })).toEqual({
      to: "/designer",
      search: { instanceId: "i-1" },
    });
  });

  test("a notification that names nothing leads nowhere", () => {
    expect(notificationTarget({})).toBeNull();
    expect(notificationTarget({ link: "" })).toBeNull();
  });

  test("a link to somewhere the router does not know is not followed", () => {
    expect(notificationTarget({ link: "https://example.com/phish" })).toBeNull();
  });
});
