import type { SessionInfo } from "@authsentinel/sdk-core";

export type { SessionInfo, SessionUser } from "@authsentinel/sdk-core";

export interface NodeMiddlewareOptions {
  /** Agent base URL (e.g. http://authsentinel-agent:8080). */
  agentBaseUrl: string;
  /** Session cookie name (must match agent config). */
  cookieName: string;
  /** Optional custom header to forward Set-Cookie from agent (default: forward to response). */
  forwardSetCookie?: boolean;
}

export interface RequestWithAuth {
  authsentinel?: {
    session: SessionInfo | null;
    sessionCookie: string | null;
  };
}

/**
 * Server-side middleware that resolves the session by calling the agent GET /session
 * with the request's session cookie, and attaches the result to the request.
 * Use with Express, Fastify, or any (req, res, next) style API.
 *
 * - Reads cookie from req.headers.cookie (or req.cookies if present).
 * - Calls agent GET /session with that cookie.
 * - Sets req.authsentinel = { session, sessionCookie }.
 * - If the agent returns Set-Cookie, the middleware can set it on the response (forwardSetCookie: true).
 */
export function createNodeMiddleware(
  options: NodeMiddlewareOptions
): (req: any, res: any, next: (err?: any) => void) => void {
  const {
    agentBaseUrl,
    cookieName,
    forwardSetCookie = true,
  } = options;
  const base = agentBaseUrl.replace(/\/$/, "");

  return function nodeMiddleware(
    req: any,
    res: any,
    next: (err?: any) => void
  ): void {
    const cookieHeader = req.headers?.cookie ?? "";
    const match = cookieHeader.match(
      new RegExp(`(?:^|;)\\s*${cookieName}=([^;]*)`)
    );
    const sessionCookie = match ? decodeURIComponent(match[1].trim()) : null;

    if (!sessionCookie) {
      (req as RequestWithAuth).authsentinel = { session: null, sessionCookie: null };
      return next();
    }

    const url = `${base}/session`;
    const fetchOpts: RequestInit = {
      method: "GET",
      headers: { Cookie: `${cookieName}=${sessionCookie}` },
    };

    fetch(url, fetchOpts)
      .then((response) => {
        if (response.status !== 200) {
          (req as RequestWithAuth).authsentinel = { session: null, sessionCookie };
          return next();
        }
        const setCookieHeader = response.headers.get("Set-Cookie");
        if (forwardSetCookie && setCookieHeader && res?.setHeader) {
          res.setHeader("Set-Cookie", setCookieHeader);
        }
        return response.json();
      })
      .then((data: { is_authenticated?: boolean; user?: SessionInfo["user"] }) => {
        const status = data?.is_authenticated ? "authenticated" : "unauthenticated";
        const session: SessionInfo = {
          status,
          user: data?.user ?? null,
        };
        (req as RequestWithAuth).authsentinel = { session, sessionCookie };
        next();
      })
      .catch((err) => next(err));
  };
}
