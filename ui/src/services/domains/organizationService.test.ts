import { afterEach, describe, expect, test } from "bun:test";
import { organizationService } from "./organizationService";
import { stubFetch } from "../shared/stubbedFetch";

describe("organizationService.createOrganization", () => {
  let restore = () => {};
  afterEach(() => restore());

  test("raises the refusal instead of returning it", async () => {
    ({ restore } = stubFetch({ error: "not permitted" }));
    await expect(organizationService.createOrganization("Acme", "")).rejects.toThrow("not permitted");
  });
});
