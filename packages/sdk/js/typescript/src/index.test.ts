import { describe, it, expect } from "vitest";
import { createClient } from "./index";

describe("typescript package", () => {
  it("re-exports createClient from core", () => {
    const client = createClient();
    expect(client.getSession).toBeDefined();
    expect(client.getLoginURL).toBeDefined();
    expect(client.getLogoutURL).toBeDefined();
  });

  it("createClient without options returns stub session unknown", async () => {
    const client = createClient();
    const session = await client.getSession();
    expect(session.status).toBe("unknown");
  });
});
