import type { Metadata } from "next";
import { SponsorDashboard } from "@/features/sponsor/components/sponsor-dashboard";

export const metadata: Metadata = {
  title: "Sponsor home - Eventory",
  description: "Find tournaments to support and manage your sponsorships.",
};

export default function SponsorDashboardPage() {
  return <SponsorDashboard />;
}
