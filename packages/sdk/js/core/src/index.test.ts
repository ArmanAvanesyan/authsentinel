import { describe, it, expect } from "vitest";
import { createClient } from "./index";

describe("createClient", () => {
  it("returns a client whose getSession resolves to status unknown", async () => {
    const client = createClient();
    const session = await client.getSession();
    expect(session.status).toBe("unknown");
  });
});

