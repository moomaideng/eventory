"use client";

import {
  ArrowLeft,
  ArrowRight,
  Check,
  CircleDashed,
  LockKeyhole,
  RefreshCw,
  X,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { statusLabel } from "@/features/organizer/utils";

export function RefreshDashboard({
  fetching,
  onRefresh,
}: {
  fetching: boolean;
  onRefresh: () => void;
}) {
  return (
    <Button
      variant="outline"
      size="icon"
      aria-label="Refresh dashboard"
      title="Refresh dashboard"
      disabled={fetching}
      onClick={onRefresh}
    >
      <RefreshCw className={cn(fetching && "motion-safe:animate-spin")} />
    </Button>
  );
}

export function DashboardStatus({ status }: { status: string }) {
  const Icon =
    status === "ACCEPTED"
      ? Check
      : status === "LOCKED"
        ? LockKeyhole
        : status === "REJECTED"
          ? X
          : CircleDashed;
  return (
    <Badge variant={status === "ACCEPTED" ? "secondary" : "outline"}>
      <Icon data-icon="inline-start" />
      <span className="capitalize">{statusLabel(status)}</span>
    </Badge>
  );
}

export function DashboardLoading() {
  return (
    <div
      className="flex flex-col gap-8"
      role="status"
      aria-label="Loading dashboard"
    >
      <Skeleton className="h-10 w-full max-w-80" />
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-36 w-full rounded-lg" />
        ))}
      </div>
      <Skeleton className="h-64 w-full rounded-lg" />
    </div>
  );
}

export function DashboardError({ onRetry }: { onRetry: () => void }) {
  return (
    <Empty className="border">
      <EmptyHeader>
        <EmptyTitle>Dashboard unavailable</EmptyTitle>
        <EmptyDescription>
          The tournament may not belong to your account, or the service may be
          temporarily unavailable.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="outline" onClick={onRetry}>
          <RefreshCw data-icon="inline-start" />
          Try again
        </Button>
      </EmptyContent>
    </Empty>
  );
}

export function DashboardPagination({
  page,
  pages,
  onPage,
  disabled = false,
}: {
  page: number;
  pages: number;
  onPage: (page: number) => void;
  disabled?: boolean;
}) {
  if (pages <= 1) return null;
  return (
    <nav
      aria-label="Dashboard pagination"
      className="flex items-center justify-center gap-4"
    >
      <Button
        variant="outline"
        size="icon"
        title="Previous page"
        aria-label="Previous page"
        disabled={page <= 1 || disabled}
        onClick={() => onPage(page - 1)}
      >
        <ArrowLeft />
      </Button>
      <span className="text-muted-foreground text-sm tabular-nums">
        Page {page} of {pages}
      </span>
      <Button
        variant="outline"
        size="icon"
        title="Next page"
        aria-label="Next page"
        disabled={page >= pages || disabled}
        onClick={() => onPage(page + 1)}
      >
        <ArrowRight />
      </Button>
    </nav>
  );
}
