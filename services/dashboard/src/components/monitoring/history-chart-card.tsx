import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { useMetric } from "@/hooks/use-metric";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from "@/components/ui/chart";
import { Skeleton } from "@/components/ui/skeleton";

export type HistoryRange = "1h" | "24h" | "7d";

/** One row per timestamp (Unix seconds), plus one numeric key per series. */
type HistoryPoint = Record<string, number>;

const timeFormat = new Intl.DateTimeFormat(undefined, {
  hour: "2-digit",
  minute: "2-digit",
});
const dayFormat = new Intl.DateTimeFormat(undefined, {
  day: "2-digit",
  month: "2-digit",
});
const fullFormat = new Intl.DateTimeFormat(undefined, {
  dateStyle: "short",
  timeStyle: "short",
});

interface HistoryChartCardProps {
  title: string;
  description: string;
  path: string;
  range: HistoryRange;
  config: ChartConfig;
  formatValue: (value: number) => string;
  yDomain?: [number, number];
}

export function HistoryChartCard({
  title,
  description,
  path,
  range,
  config,
  formatValue,
  yDomain,
}: HistoryChartCardProps) {
  const { data, loading, error } = useMetric<HistoryPoint[]>(
    `${path}?range=${range}`,
  );
  const tickFormat = range === "7d" ? dayFormat : timeFormat;

  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent>
        {loading ? (
          <Skeleton className="aspect-video w-full" />
        ) : error ? (
          <p className="text-sm text-destructive">{error}</p>
        ) : !data?.length ? (
          <p className="text-sm text-muted-foreground">
            No data for this range
          </p>
        ) : (
          <ChartContainer config={config} className="aspect-video w-full">
            <AreaChart data={data} margin={{ left: 0, right: 8 }}>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="timestamp"
                type="number"
                scale="time"
                domain={["dataMin", "dataMax"]}
                tickFormatter={(ts: number) => tickFormat.format(ts * 1000)}
                tickLine={false}
                axisLine={false}
                minTickGap={32}
              />
              <YAxis
                domain={yDomain}
                tickFormatter={formatValue}
                tickLine={false}
                axisLine={false}
                width={72}
              />
              <ChartTooltip
                content={
                  <ChartTooltipContent
                    labelFormatter={(_, payload) =>
                      fullFormat.format(payload[0]?.payload.timestamp * 1000)
                    }
                    formatter={(value, name) => (
                      <>
                        <span className="text-muted-foreground">
                          {config[String(name)]?.label ?? name}
                        </span>
                        <span className="ml-auto font-mono font-medium tabular-nums">
                          {formatValue(Number(value))}
                        </span>
                      </>
                    )}
                  />
                }
              />
              {Object.keys(config).length > 1 && (
                <ChartLegend content={<ChartLegendContent />} />
              )}
              {Object.keys(config).map((key) => (
                <Area
                  key={key}
                  dataKey={key}
                  type="monotone"
                  stroke={`var(--color-${key})`}
                  fill={`var(--color-${key})`}
                  fillOpacity={0.2}
                  strokeWidth={1.5}
                  dot={false}
                  isAnimationActive={false}
                />
              ))}
            </AreaChart>
          </ChartContainer>
        )}
      </CardContent>
    </Card>
  );
}
