import type { Metadata } from "next";
import { AccountSettingsManager } from "@/features/account/components/account-settings-manager";
import { OrganizerProfileManager } from "@/features/organizer/components/organizer-profile-manager";
import { SponsorProfileManager } from "@/features/sponsor/components/sponsor-profile-manager";

export const metadata: Metadata = {
  title: "Profile settings | Eventory",
  description: "Manage your account and your organizer and sponsor profiles.",
};

export default function SettingsPage() {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-4 py-10">
      <header className="flex flex-col gap-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Profile settings
        </h1>
        <p className="text-muted-foreground text-sm">
          Manage your account, and your organizer and sponsor profiles.
        </p>
      </header>
      <AccountSettingsManager />
      <OrganizerProfileManager />
      <SponsorProfileManager />
    </div>
  );
}
