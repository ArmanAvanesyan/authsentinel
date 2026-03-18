import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import type { AuthClient, SessionInfo } from "@authsentinel/sdk-core";
import { createClient } from "@authsentinel/sdk-core";

export type { AuthClient, AuthClientConfig, SessionInfo, SessionUser } from "@authsentinel/sdk-core";
export { createClient } from "@authsentinel/sdk-core";

export interface AuthProviderProps {
  children: ReactNode;
  /** Agent base URL. If not set, session will remain "unknown". */
  agentBaseUrl?: string;
}

export interface AuthContextValue {
  session: SessionInfo | null;
  loading: boolean;
  error: Error | null;
  client: AuthClient;
  refresh: () => Promise<void>;
  login: (returnUrl?: string) => void;
  logout: (redirectTo?: string) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider(props: AuthProviderProps) {
  const { children, agentBaseUrl } = props;
  const [session, setSession] = useState<SessionInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const client = useMemo(
    () => createClient({ agentBaseUrl }),
    [agentBaseUrl]
  );

  const loadSession = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const info = await client.getSession();
      setSession(info);
    } catch (e) {
      setError(e instanceof Error ? e : new Error(String(e)));
      setSession({ status: "unknown" });
    } finally {
      setLoading(false);
    }
  }, [client]);

  const refresh = useCallback(async () => {
    setError(null);
    try {
      await client.refresh();
      const info = await client.getSession();
      setSession(info);
    } catch (e) {
      setError(e instanceof Error ? e : new Error(String(e)));
    }
  }, [client]);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      loading,
      error,
      client,
      refresh,
      login: (returnUrl?: string) => client.login(returnUrl),
      logout: (redirectTo?: string) => client.logout(redirectTo),
    }),
    [session, loading, error, client, refresh]
  );

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}

export function useSession(): SessionInfo | null {
  return useAuth().session;
}
