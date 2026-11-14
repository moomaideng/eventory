"use client";

import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/page-header";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { useOrganizerProfile } from "@/features/organizer/hooks/use-organizer-profile";
import { OrganizerOnboardingModal } from "@/features/organizer/components/organizer-onboarding-modal";
import { SponsorDirectory } from "@/features/organizer/components/sponsor-directory";

export function SponsorDirectoryWorkspace() {
  const { hasProfile, isLoading, needsOnboarding, error, refetch } =
    useOrganizerProfile();

  // Only a confirmed missing profile should trigger setup, not an API outage.
  const profileNotFound =
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    error.status === 404;
  const showOnboarding = needsOnboarding && profileNotFound;

  // Mount the directory only once profile setup has populated the shared cache.
  if (hasProfile) return <SponsorDirectory />;

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace="organizer"
        title="Sponsor directory"
        description="Find sponsor profiles and contact them about your tournament."
      />

      {isLoading ? (
        <Skeleton className="h-40 w-full rounded-xl" />
      ) : showOnboarding ? (
        <>
          <p className="text-muted-foreground text-sm">
            Set up your organizer profile to browse sponsors.
          </p>
          <OrganizerOnboardingModal
            open
            redirectToOnCancel="/organizer"
            cancelMessage="Cancel & Return to Tournaments"
          />
        </>
      ) : (
        <Empty className="border">
          <EmptyHeader>
            <EmptyTitle>Could not load your organizer profile</EmptyTitle>
            <EmptyDescription>
              Please try again to continue to the sponsor directory.
            </EmptyDescription>
          </EmptyHeader>
          <Button variant="outline" onClick={() => void refetch()}>
            Try again
          </Button>
        </Empty>
      )}
    </div>
  );
}
