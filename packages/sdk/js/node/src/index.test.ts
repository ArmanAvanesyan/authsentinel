import { describe, it, expect, vi } from "vitest";
import { createNodeMiddleware } from "./index";

describe("createNodeMiddleware", () => {
  it("returns middleware that calls next", () => {
    const mw = createNodeMiddleware();
    const next = vi.fn();

    mw({}, {}, next);

    expect(next).toHaveBeenCalledTimes(1);
  });
});

