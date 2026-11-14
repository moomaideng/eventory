"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import {
  ArrowRight,
  CalendarDays,
  Plus,
  Trophy,
  Users,
} from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { PageHeader } from "@/components/page-header";
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
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { TournamentStatusBadge } from "@/features/tournaments/components/tournament-status-badge";
import { dashboardQueryOptions } from "@/features/organizer/utils";
import {
  organizerRoleFilterSchema,
  organizerTournamentStatusFilterSchema,
  type OrganizerRoleFilter,
  type OrganizerTournamentStatusFilter,
} from "@/features/organizer/schemas";
import {
  DashboardError,
  DashboardLoading,
  DashboardPagination,
  RefreshDashboard,
} from "@/features/organizer/components/dashboard-shared";
import { StaffRoleBadge } from "@/features/tournament-staff";

const STATUS_TABS: ReadonlyArray<{
  value: OrganizerTournamentStatusFilter;
  label: string;
}> = [
  { value: "ALL", label: "All" },
  { value: "DRAFT", label: "Draft" },
  { value: "CROWDFUNDING", label: "Crowdfunding" },
  { value: "REGISTRATION", label: "Registration" },
  { value: "ONGOING", label: "Ongoing" },
  { value: "COMPLETED", label: "Completed" },
];

const ROLE_TABS: ReadonlyArray<{
  value: OrganizerRoleFilter;
  label: string;
}> = [
  { value: "ALL", label: "All roles" },
  { value: "OWNER", label: "Owner" },
  { value: "ADMIN", label: "Admin" },
  { value: "REFEREE", label: "Referee" },
];

export function OrganizerTournaments() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] =
    useState<OrganizerTournamentStatusFilter>("ALL");
  const [roleFilter, setRoleFilter] = useState<OrganizerRoleFilter>("ALL");
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/mine",
    {
      params: { query: { page, pageSize: 12 } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { ...dashboardQueryOptions, enabled: Boolean(authorizationHeader) }
  );

  const handleStatusChange = (value: string) => {
    const parsed = organizerTournamentStatusFilterSchema.safeParse(value);
    if (!parsed.success) return;
    setStatusFilter(parsed.data);
  };

  const handleRoleChange = (value: string) => {
    const parsed = organizerRoleFilterSchema.safeParse(value);
    if (!parsed.success) return;
    setRoleFilter(parsed.data);
  };

  const filteredItems = useMemo(() => {
    const items = data?.items ?? [];
    return items.filter((item) => {
      if (statusFilter !== "ALL" && item.tournament.status !== statusFilter) {
        return false;
      }
      if (roleFilter !== "ALL" && item.role !== roleFilter) {
        return false;
      }
      return true;
    });
  }, [data?.items, roleFilter, statusFilter]);

  const loading = authLoading || !authorizationHeader || isLoading;
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace="organizer"
        breadcrumbs={[{ label: "My tournaments" }]}
        title="My tournaments"
        description={
          data && !loading
            ? `${data.total} hosted ${
                data.total === 1 ? "tournament" : "tournaments"
              }`
            : undefined
        }
        actions={
          <>
            <Button
              render={<Link href="/organizer/tournaments/new" />}
              nativeButton={false}
              size="sm"
            >
              <Plus data-icon="inline-start" />
              Host tournament
            </Button>
            <RefreshDashboard
              fetching={isFetching}
              onRefresh={() => void refetch()}
            />
          </>
        }
      />

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
          <div className="flex flex-wrap items-center justify-between gap-3">
            <Tabs value={statusFilter} onValueChange={handleStatusChange}>
              <TabsList
                aria-label="Filter tournaments by status"
                className="max-w-full justify-start overflow-x-auto"
              >
                {STATUS_TABS.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value}>
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>

            <Tabs value={roleFilter} onValueChange={handleRoleChange}>
              <TabsList
                aria-label="Filter tournaments by role"
                className="max-w-full justify-start overflow-x-auto"
              >
                {ROLE_TABS.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value}>
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
          </div>

          {filteredItems.length === 0 ? (
            <Empty className="border">
              <EmptyHeader>
                <EmptyTitle>No matching tournaments</EmptyTitle>
                <EmptyDescription>
                  {statusFilter !== "ALL" && roleFilter !== "ALL"
                    ? `You have no ${statusFilter.toLowerCase()} tournaments as ${roleFilter.toLowerCase()}.`
                    : statusFilter !== "ALL"
                      ? `You have no ${statusFilter.toLowerCase()} tournaments.`
                      : `You are not assigned as a ${roleFilter.toLowerCase()} on any tournaments.`}
                </EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setStatusFilter("ALL");
                    setRoleFilter("ALL");
                  }}
                >
                  Show all tournaments
                </Button>
              </EmptyContent>
            </Empty>
          ) : (
            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
              {filteredItems.map(
                ({ tournament, metrics, funding, role }) => (
                  <Card
                    key={tournament.id}
                    className="group motion-safe:animate-in motion-safe:fade-in-0 motion-safe:slide-in-from-bottom-2 min-w-0 rounded-lg transition-shadow duration-200 hover:shadow-md"
                  >
                    <CardHeader>
                      <div className="mb-3 flex flex-wrap items-center gap-2">
                        {role ? <StaffRoleBadge role={role} /> : null}
                        <Badge variant="secondary">{tournament.game}</Badge>
                        <TournamentStatusBadge
                          status={tournament.status}
                          registrationDeadline={tournament.registrationDeadline}
                        />
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
                            tournament.capacity > 0
                              ? `${metrics.acceptedEntries} / ${tournament.capacity}`
                              : `${metrics.acceptedEntries} entries`
                          }
                        </ProgressValue>
                      </Progress>
                      <div className="flex flex-wrap justify-between gap-3 text-sm">
                        <span className="text-muted-foreground flex items-center gap-2">
                          <Users className="size-4" />
                          {metrics.confirmedParticipants} participants
                        </span>
                        <span className="font-medium">
                          {formatMoney(
                            funding.raisedAmountMinor,
                            funding.currency
                          )}{" "}
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
