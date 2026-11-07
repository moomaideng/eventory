import type { Metadata } from "next";
import { SponsorDirectoryWorkspace } from "@/features/organizer/components/sponsor-directory-workspace";

export const metadata: Metadata = {
  title: "Sponsor directory - Eventory",
  description:
    "Find sponsors who are open to hearing from tournament organizers.",
};

export default function SponsorDirectoryPage() {
  return <SponsorDirectoryWorkspace />;
}
