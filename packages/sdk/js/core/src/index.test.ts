import { describe, it, expect } from "vitest";
import { createClient } from "./index";
import type { Principal, SessionInfo, SessionUser } from "./types";

describe("createClient", () => {
  it("returns a client whose getSession resolves to status unknown when no agentBaseUrl", async () => {
    const client = createClient();
    const session = await client.getSession();
    expect(session.status).toBe("unknown");
  });

  it("returns stub login/logout URLs when no agentBaseUrl", () => {
    const client = createClient();
    expect(client.getLoginURL()).toBe("");
    expect(client.getLogoutURL()).toBe("");
  });

  it("returns correct login/logout URLs when agentBaseUrl is set", () => {
    const client = createClient({ agentBaseUrl: "https://auth.example.com" });
    expect(client.getLoginURL("https://app.example.com")).toBe(
      "https://auth.example.com/login?redirect_to=https%3A%2F%2Fapp.example.com"
    );
    expect(client.getLogoutURL()).toBe("https://auth.example.com/logout");
  });
});

/** Contract: SessionInfo shape expected by BFF/UI (status + optional user). */
describe("SessionInfo contract", () => {
  it("has status and optional user", () => {
    const minimal: SessionInfo = { status: "unauthenticated" };
    expect(minimal.status).toBe("unauthenticated");

    const withUser: SessionInfo = {
      status: "authenticated",
      user: { sub: "user-1", email: "u@example.com" },
    };
    expect(withUser.user?.sub).toBe("user-1");
    expect(withUser.user?.email).toBe("u@example.com");
  });

  it("SessionUser has sub and optional fields", () => {
    const user: SessionUser = {
      sub: "alice",
      email: "alice@example.com",
      preferred_username: "alice",
      name: "Alice",
      roles: ["user"],
      groups: ["g1"],
      is_admin: false,
    };
    expect(user.sub).toBe("alice");
    expect(user.roles).toEqual(["user"]);
    expect(user.groups).toEqual(["g1"]);
  });
});

/** Contract: Principal shape for API/BFF (subject + optional roles, claims, etc.). */
describe("Principal contract", () => {
  it("has subject and optional scopes, roles, claims", () => {
    const p: Principal = { subject: "user-123" };
    expect(p.subject).toBe("user-123");

    const full: Principal = {
      subject: "user-1",
      scopes: ["read"],
      roles: ["admin"],
      claims: { email: "u@example.com" },
      tenant_context: { tenant_id: "t1" },
    };
    expect(full.roles).toEqual(["admin"]);
    expect(full.claims?.email).toBe("u@example.com");
  });
});

