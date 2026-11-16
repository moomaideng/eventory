"use client";

import { useState } from "react";
import { Building2 } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { Button } from "@/components/ui/button";
import { PageHeader } from "@/components/page-header";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { useOrganizerProfile } from "@/features/organizer/hooks/use-organizer-profile";
import { OrganizerOnboardingModal } from "@/features/organizer/components/organizer-onboarding-modal";
import { SponsorDirectory } from "@/features/organizer/components/sponsor-directory";

export function SponsorDirectoryWorkspace() {
  const { user, authorizationHeader, isLoading: isAuthLoading } = useAuth();
  const { hasProfile, isLoading, needsOnboarding, error, refetch } =
    useOrganizerProfile();
  const [dismissed, setDismissed] = useState(false);

  // Only a confirmed missing profile should trigger setup, not an API outage.
  const profileNotFound =
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    error.status === 404;
  const showOnboarding = needsOnboarding && profileNotFound;

  if (!isAuthLoading && !authorizationHeader) {
    return <SponsorDirectory />;
  }

  // Mount the directory only once profile setup has populated the shared cache.
  if (hasProfile) return <SponsorDirectory />;

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace="organizer"
        breadcrumbs={[{ label: "Sponsor directory" }]}
        title="Sponsor directory"
        description="Find sponsor profiles and contact them about your tournament."
      />

      {isLoading || isAuthLoading ? (
        <Skeleton className="h-40 w-full rounded-xl" />
      ) : user?.status === "ONBOARDING" ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Building2 />
            </EmptyMedia>
            <EmptyTitle>Complete your account setup first</EmptyTitle>
            <EmptyDescription>
              Finish choosing your handle, then set up your organizer profile to
              browse sponsors.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : showOnboarding ? (
        <>
          <Empty className="border">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Building2 />
              </EmptyMedia>
              <EmptyTitle>Set up your organizer profile</EmptyTitle>
              <EmptyDescription>
                Only organizers can view sponsor contact details. Create your
                organizer profile to unlock the sponsor directory.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <Button onClick={() => setDismissed(false)}>
                Set up organizer profile
              </Button>
            </EmptyContent>
          </Empty>
          <OrganizerOnboardingModal
            open={!dismissed}
            onOpenChange={(nextOpen) => {
              if (!nextOpen) setDismissed(true);
            }}
            onSuccess={() => {
              setDismissed(false);
              void refetch();
            }}
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
