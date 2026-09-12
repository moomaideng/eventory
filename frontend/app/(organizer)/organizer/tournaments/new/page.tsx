import type { Metadata } from "next";
import { HostTournamentWorkspace } from "@/features/organizer/components/host-tournament-workspace";
import { OrganizerOnboardingModal } from "@/features/organizer/components/organizer-onboarding-modal";

export const metadata: Metadata = {
  title: "Host a Tournament - Eventory",
  description:
    "Create and publish a new esports tournament with custom team limits, rules, and entry fees.",
};

export default function HostTournamentPage() {
  return (
    <>
      <HostTournamentWorkspace />
      <OrganizerOnboardingModal redirectToOnCancel="/organizer" cancelMessage="Cancel & Return to Tournaments" />
    </>
  );
}
