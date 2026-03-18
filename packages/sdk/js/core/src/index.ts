import type {
  AuthClient,
  AuthClientConfig,
  RefreshResult,
  SessionInfo,
  SessionStatus,
} from "./types";

export type {
  AuthClient,
  AuthClientConfig,
  Principal,
  RefreshResult,
  SessionInfo,
  SessionStatus,
  SessionUser,
} from "./types";

/**
 * Creates an AuthSentinel browser client. Use in SPAs or any browser context where the agent
 * sets the session cookie (same-origin or CORS with credentials).
 *
 * - Without `agentBaseUrl`: returns a stub client (getSession() -> status "unknown").
 * - With `agentBaseUrl`: getSession() and refresh() call the agent; getLoginURL/getLogoutURL
 *   build URLs for redirect-based login/logout.
 */
export function createClient(config: AuthClientConfig = {}): AuthClient {
  const base = (config.agentBaseUrl ?? "").replace(/\/$/, "");

  if (!base) {
    return {
      async getSession(): Promise<SessionInfo> {
        return { status: "unknown" };
      },
      getLoginURL(_returnUrl?: string) {
        return "";
      },
      getLogoutURL(_redirectTo?: string) {
        return "";
      },
      async refresh(): Promise<RefreshResult> {
        return { refreshed: false };
      },
      login() {},
      logout() {},
    };
  }

  return {
    async getSession(): Promise<SessionInfo> {
      try {
        const res = await fetch(`${base}/session`, { credentials: "include" });
        if (!res.ok) {
          return { status: "unauthenticated" };
        }
        const data = (await res.json()) as {
          is_authenticated?: boolean;
          user?: SessionInfo["user"];
        };
        const status: SessionStatus = data.is_authenticated
          ? "authenticated"
          : "unauthenticated";
        return { status, user: data.user ?? null };
      } catch {
        return { status: "unknown" };
      }
    },

    getLoginURL(returnUrl?: string): string {
      const q = returnUrl ? `?redirect_to=${encodeURIComponent(returnUrl)}` : "";
      return `${base}/login${q}`;
    },

    getLogoutURL(redirectTo?: string): string {
      const q = redirectTo ? `?redirect_to=${encodeURIComponent(redirectTo)}` : "";
      return `${base}/logout${q}`;
    },

    async refresh(): Promise<RefreshResult> {
      try {
        const res = await fetch(`${base}/refresh`, {
          method: "GET",
          credentials: "include",
        });
        return { refreshed: res.ok };
      } catch {
        return { refreshed: false };
      }
    },

    login(returnUrl?: string): void {
      if (typeof window !== "undefined") {
        window.location.href = this.getLoginURL(returnUrl);
      }
    },

    logout(redirectTo?: string): void {
      if (typeof window !== "undefined") {
        window.location.href = this.getLogoutURL(redirectTo);
      }
    },
  };
}
