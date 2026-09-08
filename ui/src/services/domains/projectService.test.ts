import { afterEach, describe, expect, test } from "bun:test";
import { projectService } from "./projectService";
import { stubFetch } from "../shared/stubbedFetch";

describe("projectService.createProject", () => {
  let restore = () => {};
  afterEach(() => restore());

  test("raises the refusal instead of returning it", async () => {
    ({ restore } = stubFetch({ error: "not permitted" }));
    await expect(projectService.createProject("org-1", "Payroll", "")).rejects.toThrow("not permitted");
  });

  test("hands back the created project", async () => {
    ({ restore } = stubFetch({ project: { id: "p-1", name: "Payroll" } }));
    const { project } = await projectService.createProject("org-1", "Payroll", "");
    expect(project?.id).toBe("p-1");
  });
});
