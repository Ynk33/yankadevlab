import { useCallback, useEffect, useState } from "react";

const API_BASE = import.meta.env.VITE_MONITORING_API_URL ?? "";
/** Polling period, aligned with the Prometheus scrape interval. */
const REFRESH_INTERVAL_MS = 5_000;

export function useMetric<T>(path: string) {
  const [data, setData] = useState<T | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);

  const load = useCallback(
    () =>
      fetch(`${API_BASE}${path}`, { credentials: "include" })
        .then((res) => {
          if (res.status === 401) window.location.reload();
          if (!res.ok) throw new Error(`Request failed (${res.status})`);
          return res.json();
        })
        .then((json: T) => {
          setData(json);
          setError(undefined);
        })
        .catch((err) =>
          setError(
            err instanceof Error ? err.message : "Failed to load metric",
          ),
        )
        .finally(() => setLoading(false)),
    [path],
  );

  useEffect(() => {
    load();
    const id = setInterval(load, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  const refetch = useCallback(() => {
    setLoading(true);
    setError(undefined);
    return load();
  }, [load]);

  return { data, loading, error, refetch };
}
