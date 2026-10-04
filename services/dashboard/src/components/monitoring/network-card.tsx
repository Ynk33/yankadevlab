import { ArrowDownIcon, ArrowUpIcon, NetworkIcon } from "lucide-react";
import { useMetric } from "@/hooks/use-metric";
import { MetricCard } from "@/components/monitoring/metric-card";
import { formatRate } from "@/lib/format";

interface NetworkMetric {
  rx_bytes_per_sec: number;
  tx_bytes_per_sec: number;
  timestamp: number;
}

export function NetworkCard() {
  const { data, loading, error, refetch } =
    useMetric<NetworkMetric>("/metrics/network");

  return (
    <MetricCard
      icon={NetworkIcon}
      title="Network I/O"
      description="Current traffic"
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
