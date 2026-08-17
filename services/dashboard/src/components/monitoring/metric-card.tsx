import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import { RefreshCwIcon } from "lucide-react";
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

interface MetricCardProps {
  icon: LucideIcon;
  title: string;
  description: string;
  loading: boolean;
  error: string | undefined;
  onRefresh: () => void;
  children: ReactNode;
}

export function MetricCard({
  icon: Icon,
  title,
  description,
  loading,
  error,
  onRefresh,
  children,
}: MetricCardProps) {
  return (
    <Card className="w-full max-w-xs">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Icon className="size-4" />
          {title}
        </CardTitle>
        <CardDescription>{description}</CardDescription>
        <CardAction>
          <Button
            variant="ghost"
            size="icon"
            onClick={onRefresh}
            disabled={loading}
            aria-label={`Refresh ${title}`}
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
          children
        )}
      </CardContent>
    </Card>
  );
}
