import type { Metadata } from "next";
import { OrganizerTournaments } from "@/features/organizer/components/organizer-tournaments";

export const metadata: Metadata = {
  title: "My Tournaments - Eventory",
  description: "Your hosted tournaments, registrations, and funding overview.",
};

export default function OrganizerPage() {
  return <OrganizerTournaments />;
}
