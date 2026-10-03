import { useCallback, useEffect, useState } from "react";
import { useAuth } from "@/hooks/use-auth";

const API_BASE = import.meta.env.VITE_MONITORING_API_URL ?? "";

export function useMetric<T>(path: string) {
  const { authFetch } = useAuth();
  const [data, setData] = useState<T | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);

  const load = useCallback(
    () =>
      authFetch(`${API_BASE}${path}`)
        .then((res) => {
          if (!res.ok) throw new Error(`Request failed (${res.status})`);
          return res.json();
        })
        .then((json: T) => setData(json))
        .catch((err) =>
          setError(
            err instanceof Error ? err.message : "Failed to load metric",
          ),
        )
        .finally(() => setLoading(false)),
    [authFetch, path],
  );

  useEffect(() => {
    load();
  }, [load]);

  const refetch = useCallback(() => {
    setLoading(true);
    setError(undefined);
    return load();
  }, [load]);

  return { data, loading, error, refetch };
}
