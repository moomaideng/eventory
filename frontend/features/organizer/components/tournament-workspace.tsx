"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, ArrowUpRight, Network, Trophy } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { RegistrationFormBuilder } from "@/features/registration-form/components/registration-form-builder";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  DashboardError,
  DashboardLoading,
  RefreshDashboard,
} from "@/features/organizer/components/dashboard-shared";
import { RegistrationOverview } from "@/features/organizer/components/registration-overview";
import { RegistrationTable } from "@/features/organizer/components/registration-table";
import { TournamentHomeTab } from "@/features/organizer/components/tournament-home-tab";
import { dashboardQueryOptions } from "@/features/organizer/utils";
import {
  LeaveTournamentDialog,
  StaffRoleBadge,
  TournamentStaffSection,
} from "@/features/tournament-staff";
import { SponsorshipTierSection } from "@/features/sponsorship";
import { AnnouncementSection } from "@/features/announcements";
import { TournamentStatusBadge } from "@/features/tournaments/components/tournament-status-badge";
import {
  isWorkspaceTab,
  type WorkspaceTab,
} from "@/features/organizer/workspace-tabs";

export function TournamentWorkspace({
  tournamentId,
  initialTab = "home",
}: {
  tournamentId: string;
  initialTab?: WorkspaceTab;
}) {
  const [tab, setTab] = useState<WorkspaceTab>(initialTab);
  function selectTab(next: WorkspaceTab) {
    setTab(next);
    const url = new URL(window.location.href);
    if (next === "home") url.searchParams.delete("tab");
    else url.searchParams.set("tab", next);
    window.history.replaceState(null, "", url);
  }
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
          {(() => {
            const isOwner = data.summary.role === "OWNER";
            const canReview = isOwner || data.summary.role === "ADMIN";
            const canAnnounce = canReview;
            const allowedTabs: Record<WorkspaceTab, boolean> = {
              home: true,
              registration: canReview,
              announcements: canAnnounce,
              staff: true,
              form: isOwner,
              sponsorship: isOwner,
            };
            const activeTab = allowedTabs[tab] ? tab : "home";

            return (
              <>
                <header className="flex flex-col items-start justify-between gap-5 sm:flex-row sm:items-center">
                  <div className="flex min-w-0 flex-1 flex-col gap-4">
                    <div className="flex flex-wrap items-center gap-2">
                      {data.summary.role ? (
                        <StaffRoleBadge role={data.summary.role} />
                      ) : null}
                      <Badge variant="secondary">
                        {data.summary.tournament.game}
                      </Badge>
                      <TournamentStatusBadge
                        status={data.summary.tournament.status}
                        registrationDeadline={
                          data.summary.tournament.registrationDeadline
                        }
                      />
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
                    <RefreshDashboard
                      fetching={isFetching}
                      onRefresh={() => void refetch()}
                    />
                    {!isOwner ? (
                      <LeaveTournamentDialog
                        tournamentId={tournamentId}
                        tournamentName={data.summary.tournament.name}
                      />
                    ) : null}
                    {["REGISTRATION", "ONGOING", "COMPLETED"].includes(
                      data.summary.tournament.status
                    ) ? (
                      <Button
                        variant="outline"
                        render={
                          <Link
                            href={`/organizer/tournaments/${tournamentId}/bracket`}
                          />
                        }
                        nativeButton={false}
                      >
                        <Network data-icon="inline-start" />
                        Bracket
                      </Button>
                    ) : null}
                    {data.summary.tournament.status !== "DRAFT" ? (
                      <Button
                        render={<Link href={`/tournaments/${tournamentId}`} />}
                        nativeButton={false}
                      >
                        Public page
                        <ArrowUpRight data-icon="inline-end" />
                      </Button>
                    ) : null}
                  </div>
                </header>

                <Tabs
                  value={activeTab}
                  onValueChange={(value) => {
                    if (isWorkspaceTab(value)) selectTab(value);
                  }}
                >
                  <div className="-mx-4 overflow-x-auto px-4 sm:mx-0 sm:px-0">
                  <TabsList className="w-max">
                    <TabsTrigger value="home">Home</TabsTrigger>
                    {canReview ? (
                      <TabsTrigger value="registration">
                        Registration
                      </TabsTrigger>
                    ) : null}
                    {canAnnounce ? (
                      <TabsTrigger value="announcements">
                        Announcements
                      </TabsTrigger>
                    ) : null}
                    <TabsTrigger value="staff">Staff</TabsTrigger>
                    {isOwner ? (
                      <TabsTrigger value="form">
                        Registration questions
                      </TabsTrigger>
                    ) : null}
                    {isOwner ? (
                      <TabsTrigger value="sponsorship">
                        Sponsorship
                      </TabsTrigger>
                    ) : null}
                  </TabsList>
                  </div>

                  {/* 1. Home Tab: Default landing page for all staff */}
                  <TabsContent value="home" keepMounted>
                    <TournamentHomeTab
                      tournamentId={tournamentId}
                      summary={data.summary}
                      isOwner={isOwner}
                      onRefresh={() => void refetch()}
                      onSelectTab={selectTab}
                    />
                  </TabsContent>

                  {/* 2. Registration Tab: Only for Owner & Admins */}
                  {canReview ? (
                    <TabsContent value="registration" keepMounted>
                      <div className="flex flex-col gap-6">
                        <div className="flex flex-wrap items-center justify-between gap-4">
                          <div>
                            <h2 className="text-xl font-semibold tracking-tight">
                              Registration & Applications
                            </h2>
                            <p className="text-muted-foreground text-sm">
                              Live participant pipeline, capacity, and applicant review.
                            </p>
                          </div>
                        </div>
                        <RegistrationOverview summary={data.summary} />
                        <RegistrationTable
                          key={tournamentId}
                          entries={data.entries ?? []}
                          metrics={data.summary.metrics}
                          tournamentId={tournamentId}
                          canReview={
                            canReview &&
                            data.summary.tournament.status === "REGISTRATION"
                          }
                        />
                      </div>
                    </TabsContent>
                  ) : null}

                  {/* 3. Announcements Tab */}
                  {canAnnounce ? (
                    <TabsContent value="announcements" keepMounted>
                      <AnnouncementSection
                        tournamentId={tournamentId}
                        published={data.summary.tournament.status !== "DRAFT"}
                      />
                    </TabsContent>
                  ) : null}

                  {/* 4. Staff Tab: Accessible to all tournament staff */}
                  <TabsContent value="staff" keepMounted>
                    <TournamentStaffSection
                      tournamentId={tournamentId}
                      isOwner={isOwner}
                    />
                  </TabsContent>

                  {/* 5. Registration Form Builder Tab */}
                  {isOwner ? (
                    <TabsContent value="form" keepMounted>
                      <RegistrationFormBuilder
                        tournamentId={tournamentId}
                        status={data.summary.tournament.status}
                        onViewRegistrations={
                          canReview ? () => selectTab("registration") : undefined
                        }
                      />
                    </TabsContent>
                  ) : null}

                  {/* 6. Sponsorship Tab */}
                  {isOwner ? (
                    <TabsContent value="sponsorship">
                      <SponsorshipTierSection
                        tournamentId={tournamentId}
                        currency={data.summary.tournament.currency}
                        editable={data.summary.tournament.status === "DRAFT"}
                      />
                    </TabsContent>
                  ) : null}
                </Tabs>
              </>
            );
          })()}
        </div>
      )}
    </div>
  );
}

export const TournamentDashboard = TournamentWorkspace;
