import type { Metadata } from "next";
import { AccountSettingsManager } from "@/features/account/components/account-settings-manager";

export const metadata: Metadata = {
  title: "Account settings | Eventory",
  description: "View and update your display name, handle, phone and avatar.",
};

export default function SettingsPage() {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-4 py-10">
      <header className="flex flex-col gap-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Account settings
        </h1>
        <p className="text-muted-foreground text-sm">
          Your profile is shared across competitor, organizer and sponsor modes.
        </p>
      </header>
      <AccountSettingsManager />
    </div>
  );
}
