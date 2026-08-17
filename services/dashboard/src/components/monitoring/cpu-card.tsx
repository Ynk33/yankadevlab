import { useCallback, useEffect, useState } from "react";
import { CpuIcon, RefreshCwIcon } from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

const API_BASE = import.meta.env.VITE_MONITORING_API_URL ?? "";

interface CpuMetric {
  usage_percent: number;
  timestamp: number;
}

export function CpuCard() {
  const { authFetch } = useAuth();
  const [metric, setMetric] = useState<CpuMetric | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | undefined>(undefined);

  const fetchCpu = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const res = await authFetch(`${API_BASE}/metrics/cpu`);
      if (!res.ok) throw new Error(`Request failed (${res.status})`);
      const data: CpuMetric = await res.json();
      setMetric(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load metric");
    } finally {
      setLoading(false);
    }
  }, [authFetch]);

  useEffect(() => {
    fetchCpu();
  }, [fetchCpu]);

  return (
    <Card className="w-full max-w-xs">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <CpuIcon className="size-4" />
          CPU usage
        </CardTitle>
        <CardDescription>Server-wide, 5 min average</CardDescription>
        <CardAction>
          <Button
            variant="ghost"
            size="icon"
            onClick={fetchCpu}
            disabled={loading}
            aria-label="Refresh CPU metric"
          >
            <RefreshCwIcon
              className={cn("size-4", loading && "animate-spin")}
            />
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="h-9 w-24" />
        ) : error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : (
          <p className="text-3xl font-semibold tabular-nums">
            {metric?.usage_percent.toFixed(1)}%
          </p>
        )}
      </CardContent>
    </Card>
  );
}
