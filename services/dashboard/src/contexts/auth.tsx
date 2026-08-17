import { useCallback, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { AuthContext } from "@/hooks/use-auth";

const API_BASE = import.meta.env.VITE_AUTH_API_URL ?? "";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | undefined>(
    () => localStorage.getItem("access_token") ?? undefined,
  );

  const isAuthenticated = accessToken !== undefined;

  const refreshAccessToken = useCallback(async (): Promise<
    string | undefined
  > => {
    try {
      const res = await fetch(`${API_BASE}/refresh`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) return undefined;

      const data = await res.json();
      if (!data?.access_token) return undefined;

      localStorage.setItem("access_token", data.access_token);
      setAccessToken(data.access_token);
      return data.access_token;
    } catch {
      return undefined;
    }
  }, []);

  // Try to refresh the access token on mount (if we have a refresh cookie)
  useEffect(() => {
    if (accessToken) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    refreshAccessToken();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${API_BASE}/login`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => undefined);
      throw new Error(body?.error ?? "Login failed");
    }

    const data = await res.json();
    localStorage.setItem("access_token", data.access_token);
    setAccessToken(data.access_token);
  }, []);

  const devLogin = useCallback(async (email: string, password: string) => {
    console.log(
      `Fake login with ${email} and ${password.slice(0, 2)}*****${password.slice(-2)}`,
    );
    localStorage.setItem("access_token", "dev-access-token");
    setAccessToken("dev-access-token");
  }, []);

  const logout = useCallback(async () => {
    await fetch(`${API_BASE}/logout`, {
      method: "POST",
      credentials: "include",
    }).catch(() => {
      // Best-effort — clear local state regardless
    });

    localStorage.removeItem("access_token");
    setAccessToken(undefined);
  }, []);

  const authFetch = useCallback(
    async (input: RequestInfo | URL, init: RequestInit = {}) => {
      const withAuth = (token: string | undefined): RequestInit => ({
        ...init,
        headers: { ...init.headers, Authorization: `Bearer ${token}` },
      });

      const res = await fetch(input, withAuth(accessToken));
      if (res.status !== 401) return res;

      const newToken = await refreshAccessToken();
      if (!newToken) return res;
      return fetch(input, withAuth(newToken));
    },
    [accessToken, refreshAccessToken],
  );

  const exposedLogin = useMemo(
    () => (import.meta.env.DEV ? devLogin : login),
    [devLogin, login],
  );

  const value = useMemo(
    () => ({
      accessToken,
      isAuthenticated,
      login: exposedLogin,
      logout,
      authFetch,
    }),
    [accessToken, isAuthenticated, exposedLogin, logout, authFetch],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
