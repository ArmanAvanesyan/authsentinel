export type SessionStatus = "unknown" | "authenticated" | "unauthenticated";

export interface SessionInfo {
  status: SessionStatus;
}

export function createClient(): { getSession: () => Promise<SessionInfo> } {
  // TODO: implement real browser/client runtime.
  return {
    async getSession() {
      return { status: "unknown" };
    },
  };
}

