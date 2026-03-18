import { describe, it, expect, vi } from "vitest";
import { createNodeMiddleware } from "./index";

describe("createNodeMiddleware", () => {
  it("returns middleware that calls next and sets authsentinel when no cookie", () => {
    const mw = createNodeMiddleware({
      agentBaseUrl: "http://agent.test",
      cookieName: "sess",
    });
    const req: any = { headers: {} };
    const res: any = { setHeader: vi.fn() };
    return new Promise<void>((resolve, reject) => {
      mw(req, res, (err?: any) => {
        if (err) return reject(err);
        expect(req.authsentinel).toEqual({ session: null, sessionCookie: null });
        resolve();
      });
    });
  });

  it("returns middleware that accepts options", () => {
    const mw = createNodeMiddleware({
      agentBaseUrl: "https://auth.example.com",
      cookieName: "session",
      forwardSetCookie: false,
    });
    expect(typeof mw).toBe("function");
  });
});
