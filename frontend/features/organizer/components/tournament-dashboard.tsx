"use client";

import Link from "next/link";
import { ArrowLeft, ArrowUpRight, Trophy } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DashboardError,
  DashboardLoading,
  RefreshDashboard,
} from "@/features/organizer/components/dashboard-shared";
import { DashboardOverview } from "@/features/organizer/components/dashboard-overview";
import { RegistrationTable } from "@/features/organizer/components/registration-table";
import { StatusControlDialog } from "@/features/organizer/components/status-control-dialog";
import { dashboardQueryOptions, statusLabel } from "@/features/organizer/utils";

export function TournamentDashboard({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/dashboard",
    {
      params: { path: { id: tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { ...dashboardQueryOptions, enabled: Boolean(authorizationHeader) }
  );
  const loading = authLoading || !authorizationHeader || isLoading;
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <div>
        <Button
          variant="outline"
          size="sm"
          render={<Link href="/organizer" />}
          nativeButton={false}
        >
          <ArrowLeft data-icon="inline-start" />
          My tournaments
        </Button>
      </div>
      {loading ? (
        <DashboardLoading />
      ) : error || !data ? (
        <DashboardError onRetry={() => void refetch()} />
      ) : (
        <div className="motion-safe:animate-in motion-safe:fade-in-0 flex min-w-0 flex-col gap-8 motion-safe:duration-300">
          <header className="flex flex-col items-start justify-between gap-5 sm:flex-row sm:items-center">
            <div className="flex min-w-0 flex-1 flex-col gap-4">
              <div className="flex flex-wrap gap-2">
                <Badge variant="secondary">
                  {data.summary.tournament.game}
                </Badge>
                <Badge variant="outline" className="capitalize">
                  {statusLabel(data.summary.tournament.status)}
                </Badge>
                <Badge variant="outline">
                  {data.summary.published ? "Published" : "Unpublished"}
                </Badge>
              </div>
              <div className="flex items-center gap-4">
                <div className="bg-primary/10 text-primary flex size-12 shrink-0 items-center justify-center rounded-lg">
                  <Trophy aria-hidden="true" />
                </div>
                <div className="min-w-0">
                  <h1 className="text-2xl font-bold wrap-anywhere sm:text-3xl">
                    {data.summary.tournament.name}
                  </h1>
                  <p className="text-muted-foreground mt-1 text-sm wrap-anywhere">
                    Hosted by {data.summary.tournament.organizerName}
                  </p>
                </div>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusControlDialog
                tournamentId={tournamentId}
                currentStatus={data.summary.tournament.status}
              />
              {data.summary.published ? (
                <Button
                  render={<Link href={`/tournaments/${tournamentId}`} />}
                  nativeButton={false}
                >
                  Public page
                  <ArrowUpRight data-icon="inline-end" />
                </Button>
              ) : null}
              <RefreshDashboard
                fetching={isFetching}
                onRefresh={() => void refetch()}
              />
            </div>
          </header>
          <DashboardOverview summary={data.summary} />
          <RegistrationTable
            key={tournamentId}
            entries={data.entries ?? []}
            metrics={data.summary.metrics}
          />
        </div>
      )}
    </div>
  );
}
