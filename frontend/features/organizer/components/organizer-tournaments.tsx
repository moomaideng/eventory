"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  CalendarDays,
  LayoutDashboard,
  Trophy,
  Users,
} from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button, buttonVariants } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import {
  Empty,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
  EmptyDescription,
  EmptyContent,
} from "@/components/ui/empty";
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import {
  formatTournamentDate,
  formatMoney,
} from "@/features/tournaments/utils";
import { dashboardQueryOptions, statusLabel } from "@/features/organizer/utils";
import {
  DashboardError,
  DashboardLoading,
  DashboardPagination,
  RefreshDashboard,
} from "@/features/organizer/components/dashboard-shared";

export function OrganizerTournaments() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const [page, setPage] = useState(1);
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/organizer/tournaments",
    {
      params: { query: { page, pageSize: 12 } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { ...dashboardQueryOptions, enabled: Boolean(authorizationHeader) }
  );

  const loading = authLoading || !authorizationHeader || isLoading;
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <header className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-4">
          <div className="bg-primary/10 text-primary flex size-12 shrink-0 items-center justify-center rounded-lg">
            <LayoutDashboard aria-hidden="true" />
          </div>
          <div className="flex min-w-0 flex-col gap-1">
            <p className="text-muted-foreground text-xs font-medium uppercase">
              Organizer workspace
            </p>
            <h1 className="text-2xl font-bold sm:text-3xl">My tournaments</h1>
            {data && !loading ? (
              <p className="text-muted-foreground text-sm">
                {data.total} hosted{" "}
                {data.total === 1 ? "tournament" : "tournaments"}
              </p>
            ) : null}
          </div>
        </div>
        <RefreshDashboard
          fetching={isFetching || loading}
          onRefresh={() => void refetch()}
        />
      </header>

      {loading ? (
        <DashboardLoading />
      ) : error ? (
        <DashboardError onRetry={() => void refetch()} />
      ) : !data?.items?.length ? (
        <Empty className="bg-muted/20 border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Trophy />
            </EmptyMedia>
            <EmptyTitle>
              {page > 1 ? "No tournaments on this page" : "No tournaments yet"}
            </EmptyTitle>
            <EmptyDescription>
              {page > 1
                ? "Return to the first page to see your current tournaments."
                : "Your hosted tournaments will appear here."}
            </EmptyDescription>
          </EmptyHeader>
          {page > 1 ? (
            <EmptyContent>
              <Button variant="outline" onClick={() => setPage(1)}>
                First page
              </Button>
            </EmptyContent>
          ) : null}
        </Empty>
      ) : (
        <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {data.items.map(({ tournament, metrics, funding, published }) => (
            <Card
              key={tournament.id}
              className="group motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-bottom-2 min-w-0 rounded-lg transition-shadow duration-200 hover:shadow-md"
            >
              <CardHeader>
                <div className="mb-3 flex flex-wrap items-center gap-2">
                  <Badge variant="secondary">{tournament.game}</Badge>
                  <Badge variant="outline" className="capitalize">
                    {statusLabel(tournament.status)}
                  </Badge>
                  {!published ? (
                    <Badge variant="outline">Unpublished</Badge>
                  ) : null}
                </div>
                <CardTitle>
                  <h2 className="text-lg font-semibold wrap-anywhere">
                    {tournament.name}
                  </h2>
                </CardTitle>
                <CardDescription>{tournament.organizerName}</CardDescription>
              </CardHeader>
              <CardContent className="flex-1 gap-5">
                <p className="text-muted-foreground flex items-start gap-2 text-sm">
                  <CalendarDays className="mt-0.5 size-4 shrink-0" />
                  {formatTournamentDate(tournament.startAt)}
                </p>
                <Progress
                  value={
                    tournament.capacity > 0
                      ? Math.min(
                          100,
                          (metrics.acceptedEntries / tournament.capacity) * 100
                        )
                      : 0
                  }
                >
                  <ProgressLabel>Accepted entries</ProgressLabel>
                  <ProgressValue>
                    {() =>
                      `${metrics.acceptedEntries} / ${tournament.capacity}`
                    }
                  </ProgressValue>
                </Progress>
                <div className="flex flex-wrap justify-between gap-3 text-sm">
                  <span className="text-muted-foreground flex items-center gap-2">
                    <Users className="size-4" />
                    {metrics.confirmedParticipants} participants
                  </span>
                  <span className="font-medium">
                    {formatMoney(funding.raisedAmount, funding.currency)} raised
                  </span>
                </div>
              </CardContent>
              <CardFooter>
                <Link
                  href={`/organizer/tournaments/${tournament.id}`}
                  aria-label={`Open ${tournament.name} dashboard`}
                  className={cn(
                    buttonVariants({ variant: "outline" }),
                    "w-full justify-between"
                  )}
                >
                  Open dashboard
                  <ArrowUpRight data-icon="inline-end" />
                </Link>
              </CardFooter>
            </Card>
          ))}
        </div>
      )}
      {!loading && !error && data ? (
        <DashboardPagination
          page={page}
          pages={Math.max(1, Math.ceil(data.total / data.pageSize))}
          onPage={setPage}
          disabled={isFetching}
        />
      ) : null}
    </div>
  );
}
