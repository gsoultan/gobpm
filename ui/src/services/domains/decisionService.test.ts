import { afterEach, describe, expect, test } from "bun:test";
import { decisionService } from "./decisionService";
import { stubFetch } from "../shared/stubbedFetch";

describe("decisionService.createDecision", () => {
  let restore = () => {};
  afterEach(() => restore());

  test("raises the refusal instead of returning it", async () => {
    ({ restore } = stubFetch({ err: "a decision with this key exists" }));
    await expect(
      decisionService.createDecision({ key: "approval-level", name: "Approval level", version: 1, inputs: [], outputs: [], rules: [] }),
    ).rejects.toThrow("a decision with this key exists");
  });
});
