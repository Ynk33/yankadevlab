import { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/hooks/use-auth";

const API_BASE = import.meta.env.VITE_MONITORING_API_URL ?? "";

export function useMetric<T>(path: string) {
  const { authFetch } = useAuth();
  const [data, setData] = useState<T | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);

  const refetch = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const res = await authFetch(`${API_BASE}${path}`);
      if (!res.ok) throw new Error(`Request failed (${res.status})`);
      const json: T = await res.json();
      setData(json);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load metric");
    } finally {
      setLoading(false);
    }
  }, [authFetch, path]);

  useEffect(() => {
    refetch();
  }, [refetch]);

  return { data, loading, error, refetch };
}
