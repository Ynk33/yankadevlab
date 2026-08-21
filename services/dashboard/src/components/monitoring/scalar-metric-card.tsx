import type { LucideIcon } from "lucide-react";
import { useMetric } from "@/hooks/use-metric";
import { MetricCard } from "@/components/monitoring/metric-card";

interface ScalarMetric {
  usage_percent: number;
  timestamp: number;
}

interface ScalarMetricCardProps {
  icon: LucideIcon;
  title: string;
  description: string;
  path: string;
}

export function ScalarMetricCard({
  icon,
  title,
  description,
  path,
}: ScalarMetricCardProps) {
  const { data, loading, error, refetch } = useMetric<ScalarMetric>(path);

  return (
    <MetricCard
      icon={icon}
      title={title}
      description={description}
      loading={loading}
      error={error}
      onRefresh={refetch}
    >
      <p className="text-3xl font-semibold tabular-nums">
        {data?.usage_percent.toFixed(1)}%
      </p>
    </MetricCard>
  );
}
