import React from "react";
import { SponsorNavbar } from "@/components/navbar/sponsor-navbar";
import { SponsorOnboardingModal } from "@/features/sponsor/components/sponsor-onboarding-modal";

export default function SponsorLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <SponsorNavbar />
      <main className="flex flex-1 flex-col">
        {children}
        <SponsorOnboardingModal redirectToOnCancel="/hub" cancelMessage="Cancel & Return to Hub" />
      </main>
    </>
  );
}
