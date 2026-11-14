import type { Metadata } from "next";
import { PageHeader } from "@/components/page-header";
import { AccountSettingsManager } from "@/features/account/components/account-settings-manager";
import { OrganizerProfileManager } from "@/features/organizer/components/organizer-profile-manager";
import { SponsorProfileManager } from "@/features/sponsor/components/sponsor-profile-manager";

export const metadata: Metadata = {
  title: "Profile settings | Eventory",
  description: "Manage your account and your organizer and sponsor profiles.",
};

export default function SettingsPage() {
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        breadcrumbs={[{ label: "Profile settings" }]}
        title="Profile settings"
        description="Manage your account, and your organizer and sponsor profiles."
      />
      <div className="flex w-full max-w-3xl flex-col gap-8">
        <AccountSettingsManager />
        <OrganizerProfileManager />
        <SponsorProfileManager />
      </div>
    </div>
  );
}
