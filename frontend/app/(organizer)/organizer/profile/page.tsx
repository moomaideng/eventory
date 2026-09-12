import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { OrganizerProfileManager } from "@/features/organizer/components/organizer-profile-manager";

export const metadata: Metadata = {
  title: "Organizer Profile - Eventory",
  description:
    "Manage the organizer name and contact email sponsors and participants see.",
};

export default function OrganizerProfilePage() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Button
        variant="ghost"
        size="sm"
        render={<Link href="/organizer" />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Organizer dashboard
      </Button>

      <OrganizerProfileManager />
    </div>
  );
}
