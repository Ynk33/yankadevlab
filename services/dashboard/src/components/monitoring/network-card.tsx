import { ArrowDownIcon, ArrowUpIcon, NetworkIcon } from "lucide-react";
import { useMetric } from "@/hooks/use-metric";
import { MetricCard } from "@/components/monitoring/metric-card";

interface NetworkMetric {
  rx_bytes_per_sec: number;
  tx_bytes_per_sec: number;
  timestamp: number;
}

function formatRate(bytesPerSec: number): string {
  if (bytesPerSec < 1024) return `${bytesPerSec.toFixed(0)} B/s`;
  const kb = bytesPerSec / 1024;
  if (kb < 1024) return `${kb.toFixed(1)} KB/s`;
  return `${(kb / 1024).toFixed(1)} MB/s`;
}

export function NetworkCard() {
  const { data, loading, error, refetch } =
    useMetric<NetworkMetric>("/metrics/network");

  return (
    <MetricCard
      icon={NetworkIcon}
      title="Network I/O"
      description="Traffic, 5 min average"
      loading={loading}
      error={error}
      onRefresh={refetch}
    >
      <div className="flex flex-col gap-1">
        <p className="flex items-center gap-1.5 text-xl font-semibold tabular-nums">
          <ArrowDownIcon className="size-4 text-muted-foreground" />
          {data ? formatRate(data.rx_bytes_per_sec) : "—"}
        </p>
        <p className="flex items-center gap-1.5 text-xl font-semibold tabular-nums">
          <ArrowUpIcon className="size-4 text-muted-foreground" />
          {data ? formatRate(data.tx_bytes_per_sec) : "—"}
        </p>
      </div>
    </MetricCard>
  );
}
