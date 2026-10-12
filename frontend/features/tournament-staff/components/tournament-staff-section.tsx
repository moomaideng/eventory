"use client";

import { useMemo } from "react";
import { AlertCircle } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { InviteStaffDialog } from "@/features/tournament-staff/components/invite-staff-dialog";
import { StaffTable } from "@/features/tournament-staff/components/staff-table";
import { sortStaffMembers } from "@/features/tournament-staff/utils";

export function TournamentStaffSection({
  tournamentId,
  isOwner,
}: {
  tournamentId: string;
  isOwner: boolean;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();

  const { data, error, isLoading, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/staff",
    {
      params: { path: { id: tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { enabled: Boolean(authorizationHeader) }
  );

  const sortedStaff = useMemo(
    () => sortStaffMembers(data?.items ?? []),
    [data?.items]
  );

  const loading = authLoading || !authorizationHeader || isLoading;

  return (
    <section className="flex min-w-0 flex-col gap-6" aria-labelledby="staff-heading">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 flex-col gap-1">
          <div className="flex items-center gap-2.5">
            <h2 id="staff-heading" className="text-xl font-bold tracking-tight">
              Tournament Staff
            </h2>
            {!loading ? (
              <Badge variant="secondary" className="rounded-full px-2.5">
                {sortedStaff.length} {sortedStaff.length === 1 ? "member" : "members"}
              </Badge>
            ) : null}
          </div>
          <p className="text-muted-foreground text-sm">
            Manage administrators and referees helping run this tournament.
          </p>
        </div>
        {isOwner && !loading ? (
          <div className="shrink-0">
            <InviteStaffDialog tournamentId={tournamentId} />
          </div>
        ) : null}
      </div>

      {loading ? (
        <div className="flex flex-col gap-3">
          <Skeleton className="h-10 w-full rounded-lg" />
          <Skeleton className="h-32 w-full rounded-xl" />
        </div>
      ) : error ? (
        <Alert variant="destructive">
          <AlertCircle />
          <AlertTitle>Unable to load staff members</AlertTitle>
          <AlertDescription className="flex items-center justify-between gap-4">
            <span>Something went wrong while loading the staff roster.</span>
            <Button variant="outline" size="sm" onClick={() => void refetch()}>
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      ) : (
        <StaffTable
          staff={sortedStaff}
          tournamentId={tournamentId}
          isOwner={isOwner}
        />
      )}
    </section>
  );
}
