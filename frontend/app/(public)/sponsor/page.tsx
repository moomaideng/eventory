import React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Gamepad2 } from "lucide-react";

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
      <div className="flex flex-col gap-4 border-t pt-6">
        <div>
          <span className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Workspace Blueprint
          </span>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Sponsor Dashboard (US5-5)
          </h1>
        </div>

        <div className="bg-muted/30 text-muted-foreground flex flex-col gap-3 rounded-xl border p-5 text-sm leading-relaxed">
          <p className="text-foreground font-semibold">
            Sprint 1 / Epic 5 Team &amp; PO Notice:
          </p>
          <p>
            This page is a placeholder blueprint scaffold to eliminate 404
            errors from Mode Hub navigation. The full feature set is being
            developed by the Sponsor domain team:
          </p>
          <ul className="list-disc space-y-1 pl-5">
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
        </div>
      </div>
    </div>
  );
}
