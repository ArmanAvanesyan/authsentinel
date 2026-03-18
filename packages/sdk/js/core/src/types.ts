/**
 * Shared principal/session types aligned with AuthSentinel proto sdk/v1 and agent API.
 */

export type SessionStatus = "unknown" | "authenticated" | "unauthenticated";

export interface Principal {
  subject: string;
  scopes?: string[];
  roles?: string[];
  claims?: Record<string, unknown>;
  tenant_context?: Record<string, unknown>;
  access_token?: string;
  expires_at?: number;
}

export interface SessionUser {
  sub: string;
  email?: string;
  preferred_username?: string;
  name?: string;
  roles?: string[];
  groups?: string[];
  is_admin?: boolean;
  tenant_context?: Record<string, unknown>;
  claims?: Record<string, unknown>;
}

export interface SessionInfo {
  status: SessionStatus;
  user?: SessionUser | null;
}

export interface RefreshResult {
  refreshed: boolean;
}

export interface AuthClientConfig {
  /** Agent base URL (e.g. https://auth.example.com). Required for real session/refresh/logout. */
  agentBaseUrl?: string;
}

export interface AuthClient {
  getSession(): Promise<SessionInfo>;
  getLoginURL(returnUrl?: string): string;
  getLogoutURL(redirectTo?: string): string;
  refresh(): Promise<RefreshResult>;
  /** Navigate to agent login (redirect). */
  login(returnUrl?: string): void;
  /** Navigate to agent logout (redirect). */
  logout(redirectTo?: string): void;
}
