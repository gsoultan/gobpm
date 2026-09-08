import { afterEach, describe, expect, test } from "bun:test";
import { definitionService } from "./definitionService";
import { stubFetch } from "../shared/stubbedFetch";

describe("definitionService.createDefinition", () => {
  let restore = () => {};
  afterEach(() => restore());

  test("raises the refusal instead of returning it", async () => {
    ({ restore } = stubFetch({ error: "definition key already deployed" }));
    await expect(
      definitionService.createDefinition("p-1", { key: "expense", name: "Expense", nodes: [], flows: [] }),
    ).rejects.toThrow("definition key already deployed");
  });
});
