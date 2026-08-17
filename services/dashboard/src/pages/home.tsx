import { CpuCard } from "@/components/monitoring/cpu-card";

export default function HomePage() {
  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      <h1 className="text-2xl font-semibold">Welcome to YankaDevLab</h1>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <CpuCard />
      </div>
    </div>
  );
}
