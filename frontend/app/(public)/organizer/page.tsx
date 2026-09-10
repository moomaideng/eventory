import React from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { ArrowLeft, Info } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

export const metadata: Metadata = {
  title: "Organizer Hub - Eventory",
  description: "Organizer dashboard and tournament management workspace.",
};

export default function OrganizerHubPage() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      {/* Navigation */}
      <Button
        variant="ghost"
        size="sm"
        render={<Link href="/hub" />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Back to Mode Hub
      </Button>

      {/* Blueprint Content */}
      <div className="flex flex-col gap-6">
        <Separator />

        <div className="flex flex-col gap-1">
          <span className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Workspace Blueprint
          </span>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Organizer Hub (US3-2)
          </h1>
        </div>

        <Alert>
          <Info />
          <AlertTitle>Sprint 1 / Epic 3 Team &amp; PO Notice</AlertTitle>
          <AlertDescription className="flex flex-col gap-3">
            <p>
              This page is a placeholder blueprint scaffold to eliminate 404
              errors from Mode Hub navigation. The full feature set is being
              developed by the Organizer domain team:
            </p>
            <ul className="flex list-disc flex-col gap-1.5 pl-5">
              <li>
                <strong>US3-2 (Organizer Overview):</strong> Dashboard
                summarizing hosted tournaments, live participant counts, and
                active funding campaigns.
              </li>
              <li>
                <strong>US3-1 (Create Tournament Wizard):</strong> 4-step wizard
                for general info, rules, crowdfunding targets, and registration
                questions.
              </li>
              <li>
                <strong>US3-3 (Status Controller):</strong> Lifecycle state
                transitions (Draft &rarr; Crowdfunding &rarr; Open &rarr; Live
                &rarr; Completed).
              </li>
              <li>
                <strong>US3-7 (Staff Management):</strong> Invite and assign
                matchday referees and administrators.
              </li>
            </ul>
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
