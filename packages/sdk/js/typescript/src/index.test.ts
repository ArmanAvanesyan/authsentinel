import { describe, it, expect } from "vitest";
import { createTypescriptHelpers } from "./index";

describe("createTypescriptHelpers", () => {
  it("returns an object (smoke test)", () => {
    const helpers = createTypescriptHelpers();
    expect(helpers).toBeTypeOf("object");
  });
});

