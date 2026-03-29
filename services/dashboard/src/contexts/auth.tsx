import {
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import type { ReactNode } from "react";
import { AuthContext } from "@/hooks/use-auth";

const API_BASE = import.meta.env.VITE_AUTH_API_URL ?? "";

export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setAccessToken] = useState<string | null>(() =>
    localStorage.getItem("access_token"),
  );

  const isAuthenticated = accessToken !== null;

  // Try to refresh the access token on mount (if we have a refresh cookie)
  useEffect(() => {
    if (accessToken) return;

    fetch(`${API_BASE}/refresh`, {
      method: "POST",
      credentials: "include",
    })
      .then((res) => {
        if (!res.ok) return null;
        return res.json();
      })
      .then((data) => {
        if (data?.access_token) {
          localStorage.setItem("access_token", data.access_token);
          setAccessToken(data.access_token);
        }
      })
      .catch(() => {
        // No valid refresh token — user needs to log in
      });
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${API_BASE}/login`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const body = await res.json().catch(() => null);
      throw new Error(body?.error ?? "Login failed");
    }

    const data = await res.json();
    localStorage.setItem("access_token", data.access_token);
    setAccessToken(data.access_token);
  }, []);

  const logout = useCallback(async () => {
    await fetch(`${API_BASE}/logout`, {
      method: "POST",
      credentials: "include",
    }).catch(() => {
      // Best-effort — clear local state regardless
    });

    localStorage.removeItem("access_token");
    setAccessToken(null);
  }, []);

  const value = useMemo(
    () => ({ accessToken, isAuthenticated, login, logout }),
    [accessToken, isAuthenticated, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
