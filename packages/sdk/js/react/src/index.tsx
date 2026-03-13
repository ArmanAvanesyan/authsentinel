import type { ReactNode } from "react";

export interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider(props: AuthProviderProps) {
  // TODO: wire to core client.
  return props.children;
}

