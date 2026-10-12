"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  ArrowRight,
  CalendarDays,
  LayoutDashboard,
  Plus,
  Trophy,
  Users,
} from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
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
import { StaffRoleBadge } from "@/features/tournament-staff";
import { cn } from "@/lib/utils";

type RoleFilter = "ALL" | "OWNER" | "ADMIN" | "REFEREE";

export function OrganizerTournaments() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const [page, setPage] = useState(1);
  const [roleFilter, setRoleFilter] = useState<RoleFilter>("ALL");
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/mine",
    {
      params: { query: { page, pageSize: 12 } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { ...dashboardQueryOptions, enabled: Boolean(authorizationHeader) }
  );

  const counts = useMemo(() => {
    const items = data?.items ?? [];
    const c = { ALL: items.length, OWNER: 0, ADMIN: 0, REFEREE: 0 };
    for (const item of items) {
      if (item.role === "OWNER") c.OWNER++;
      else if (item.role === "ADMIN") c.ADMIN++;
      else if (item.role === "REFEREE") c.REFEREE++;
    }
    return c;
  }, [data?.items]);

  const filteredItems = useMemo(() => {
    const items = data?.items ?? [];
    if (roleFilter === "ALL") return items;
    return items.filter((item) => item.role === roleFilter);
  }, [data?.items, roleFilter]);

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
        <div className="flex items-center gap-2">
          <Button
            render={<Link href="/organizer/tournaments/new" />}
            nativeButton={false}
            size="sm"
          >
            <Plus data-icon="inline-start" />
            Host tournament
          </Button>
          <RefreshDashboard
            fetching={isFetching || loading}
            onRefresh={() => void refetch()}
          />
        </div>
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
                : "Get started by hosting your first esports tournament."}
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            {page > 1 ? (
              <Button variant="outline" onClick={() => setPage(1)}>
                First page
              </Button>
            ) : (
              <Button
                render={<Link href="/organizer/tournaments/new" />}
                nativeButton={false}
              >
                <Plus data-icon="inline-start" />
                Host a tournament
              </Button>
            )}
          </EmptyContent>
        </Empty>
      ) : (
        <div className="flex flex-col gap-6">
          <div className="flex flex-wrap items-center gap-2">
            {(
              [
                { id: "ALL", label: "All" },
                { id: "OWNER", label: "Owner" },
                { id: "ADMIN", label: "Admin" },
                { id: "REFEREE", label: "Referee" },
              ] as const
            ).map(({ id, label }) => {
              const count = counts[id];
              const isActive = roleFilter === id;
              return (
                <Button
                  key={id}
                  variant={isActive ? "default" : "outline"}
                  size="sm"
                  onClick={() => setRoleFilter(id)}
                  className="rounded-full"
                >
                  {label}
                  <span
                    className={cn(
                      "ml-1.5 text-xs font-normal",
                      isActive
                        ? "text-primary-foreground/80"
                        : "text-muted-foreground"
                    )}
                  >
                    {count}
                  </span>
                </Button>
              );
            })}
          </div>

          {filteredItems.length === 0 ? (
            <Empty className="border">
              <EmptyHeader>
                <EmptyTitle>No {roleFilter.toLowerCase()} tournaments</EmptyTitle>
                <EmptyDescription>
                  You are not assigned as a {roleFilter.toLowerCase()} on any
                  tournaments.
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setRoleFilter("ALL")}
                >
                  Show all tournaments
                </Button>
              </EmptyContent>
            </Empty>
          ) : (
            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
              {filteredItems.map(
                ({ tournament, metrics, funding, published, role }) => (
                  <Card
                    key={tournament.id}
                    className="group motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-bottom-2 min-w-0 rounded-lg transition-shadow duration-200 hover:shadow-md"
                  >
                    <CardHeader>
                      <div className="mb-3 flex flex-wrap items-center gap-2">
                        {role ? <StaffRoleBadge role={role} /> : null}
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
                                (metrics.acceptedEntries / tournament.capacity) *
                                  100
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
                          {formatMoney(funding.raisedAmount, funding.currency)}{" "}
                          raised
                        </span>
                      </div>
                    </CardContent>
                    <CardFooter>
                      <Button
                        variant="outline"
                        render={
                          <Link href={`/organizer/tournaments/${tournament.id}`} />
                        }
                        nativeButton={false}
                        aria-label={`Manage ${tournament.name}`}
                        className="w-full justify-between"
                      >
                        Manage
                        <ArrowRight data-icon="inline-end" />
                      </Button>
                    </CardFooter>
                  </Card>
                )
              )}
            </div>
          )}
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
