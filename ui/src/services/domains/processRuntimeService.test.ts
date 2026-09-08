import { afterEach, describe, expect, test } from "bun:test";
import { processRuntimeService } from "./processRuntimeService";
import { stubFetch } from "../shared/stubbedFetch";

describe("processRuntimeService.startProcess", () => {
  let restore = () => {};
  afterEach(() => restore());

  test("raises the refusal instead of returning it", async () => {
    ({ restore } = stubFetch({ error: "no deployed definition with key expense" }));
    await expect(processRuntimeService.startProcess("p-1", "expense")).rejects.toThrow(
      "no deployed definition with key expense",
    );
  });

  test("hands back the instance that was started", async () => {
    ({ restore } = stubFetch({ instanceId: "i-1" }));
    const { instance_id } = await processRuntimeService.startProcess("p-1", "expense");
    expect(instance_id).toBe("i-1");
  });
});
