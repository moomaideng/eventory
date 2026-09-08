import React from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";

export default function OrganizerHubPage() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      {/* Navigation */}
      <Link
        href="/hub"
        className="text-muted-foreground hover:text-foreground inline-flex w-fit items-center gap-2 text-sm font-medium transition-colors"
      >
        <ArrowLeft className="size-4" />
        Back to Mode Hub
      </Link>

      {/* Blueprint Content */}
      <div className="flex flex-col gap-4 border-t pt-6">
        <div>
          <span className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
            Workspace Blueprint
          </span>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            Organizer Hub (US3-2)
          </h1>
        </div>

        <div className="bg-muted/30 text-muted-foreground flex flex-col gap-3 rounded-xl border p-5 text-sm leading-relaxed">
          <p className="text-foreground font-semibold">
            Sprint 1 / Epic 3 Team &amp; PO Notice:
          </p>
          <p>
            This page is a placeholder blueprint scaffold to eliminate 404
            errors from Mode Hub navigation. The full feature set is being
            developed by the Organizer domain team:
          </p>
          <ul className="list-disc space-y-1 pl-5">
            <li>
              <strong>US3-2 (Organizer Overview):</strong> Dashboard summarizing
              hosted tournaments, live participant counts, and active funding
              campaigns.
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
        </div>
      </div>
    </div>
  );
}
