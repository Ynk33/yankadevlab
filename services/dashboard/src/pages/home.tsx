import { CpuIcon, HardDriveIcon, MemoryStickIcon } from "lucide-react";
import { ScalarMetricCard } from "@/components/monitoring/scalar-metric-card";
import { NetworkCard } from "@/components/monitoring/network-card";

export default function HomePage() {
  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Welcome to YankaDevLab</h1>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <ScalarMetricCard
          icon={CpuIcon}
          title="CPU usage"
          description="Server-wide, 5 min average"
          path="/metrics/cpu"
        />
        <ScalarMetricCard
          icon={MemoryStickIcon}
          title="RAM usage"
          description="Used vs total memory"
          path="/metrics/ram"
        />
        <ScalarMetricCard
          icon={HardDriveIcon}
          title="Disk usage"
          description="Root filesystem"
          path="/metrics/disk"
        />
        <NetworkCard />
      </div>
    </div>
  );
}
