import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { ChartConfig } from "@/components/ui/chart";
import {
  HistoryChartCard,
  type HistoryRange,
} from "@/components/monitoring/history-chart-card";
import { formatRate } from "@/lib/format";

const RANGES: HistoryRange[] = ["1h", "24h", "7d"];

const usageConfig = {
  value: { label: "Usage", color: "var(--primary)" },
} satisfies ChartConfig;

const networkConfig = {
  rx: { label: "Download", color: "var(--primary)" },
  tx: { label: "Upload", color: "var(--chart-2)" },
} satisfies ChartConfig;

const formatPercent = (value: number) => `${value.toFixed(1)}%`;

export function HistorySection() {
  const [range, setRange] = useState<HistoryRange>("1h");

  return (
    <section className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-lg font-semibold">History</h2>
        <div className="flex gap-1">
          {RANGES.map((r) => (
            <Button
              key={r}
              size="sm"
              variant={r === range ? "secondary" : "ghost"}
              onClick={() => setRange(r)}
              aria-pressed={r === range}
            >
              {r}
            </Button>
          ))}
        </div>
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        <HistoryChartCard
          title="CPU usage"
          description="Server-wide"
          path="/metrics/cpu/history"
          range={range}
          config={usageConfig}
          formatValue={formatPercent}
          yDomain={[0, 100]}
        />
        <HistoryChartCard
          title="RAM usage"
          description="Used vs total memory"
          path="/metrics/ram/history"
          range={range}
          config={usageConfig}
          formatValue={formatPercent}
          yDomain={[0, 100]}
        />
        <HistoryChartCard
          title="Disk usage"
          description="Root filesystem"
          path="/metrics/disk/history"
          range={range}
          config={usageConfig}
          formatValue={formatPercent}
          yDomain={[0, 100]}
        />
        <HistoryChartCard
          title="Network I/O"
          description="Download and upload rates"
          path="/metrics/network/history"
          range={range}
          config={networkConfig}
          formatValue={formatRate}
        />
      </div>
    </section>
  );
}
