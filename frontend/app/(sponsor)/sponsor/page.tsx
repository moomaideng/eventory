import React from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ArrowLeft, Gamepad2, Info } from "lucide-react";

export const metadata: Metadata = {
  title: "Sponsor Dashboard - Eventory",
  description:
    "Sponsor campaigns, prize pool pledges, and brand placement workspace.",
};

export default function SponsorDashboardPage() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      {/* Navigation */}
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          render={<Link href="/hub" />}
          nativeButton={false}
        >
          <ArrowLeft data-icon="inline-start" />
          Back to Mode Hub
        </Button>
        <Button
          variant="ghost"
          size="sm"
          render={<Link href="/tournaments" />}
          nativeButton={false}
        >
          <Gamepad2 data-icon="inline-start" />
          Browse Tournaments
        </Button>
      </div>

      {/* Blueprint Content */}
      <div className="flex flex-col gap-6">
        <Separator />

        <div className="flex flex-col gap-1">
          <span className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Workspace Blueprint
          </span>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Sponsor Dashboard (US5-5)
          </h1>
        </div>

        <Alert>
          <Info />
          <AlertTitle>Sprint 1 / Epic 5 Team &amp; PO Notice</AlertTitle>
          <AlertDescription className="flex flex-col gap-3">
            <p>
              This page is a placeholder blueprint scaffold to eliminate 404
              errors from Mode Hub navigation. The full feature set is being
              developed by the Sponsor domain team:
            </p>
            <ul className="flex list-disc flex-col gap-1.5 pl-5">
              <li>
                <strong>US5-5 (Campaigns &amp; Pledges Overview):</strong> Summary
                of pledged prize pool tiers and campaign statuses.
              </li>
              <li>
                <strong>Brand Asset Showcase:</strong> Logo placement preview,
                company profile, and official website redirect link.
              </li>
              <li>
                <strong>Sponsor Tier Pledge Checkout:</strong> Funding integration
                to contribute to active tournament crowdfunding goals.
              </li>
            </ul>
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
