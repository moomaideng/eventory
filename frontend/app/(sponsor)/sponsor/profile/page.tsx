import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SponsorProfileForm } from "@/features/sponsor/components/sponsor-profile-form";

export const metadata: Metadata = {
  title: "Sponsor Profile - Eventory",
  description: "Manage the sponsor name and contact email organizers see.",
};

export default function SponsorProfilePage() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Button
        variant="ghost"
        size="sm"
        render={<Link href="/sponsor" />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Sponsor dashboard
      </Button>

      <SponsorProfileForm />
    </div>
  );
}
