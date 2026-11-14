"use client";

import React, { useState } from "react";
import { PageHeader } from "@/components/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { useOrganizerProfile } from "../hooks/use-organizer-profile";
import { HostTournamentForm } from "./host-tournament-form";
import { HostTournamentPreview } from "./host-tournament-preview";
import type { CreateTournamentFormInput } from "../schemas";

export function HostTournamentWorkspace() {
  const { profile, isLoading } = useOrganizerProfile();

  const [draft, setDraft] = useState<CreateTournamentFormInput>({
    name: "",
    game: "",
    location: "Online",
    description: "",
    detail: "",
  });

  if (isLoading) {
    return (
      <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
        <div className="flex flex-col gap-3">
          <Skeleton className="h-3 w-48" />
          <Skeleton className="h-9 w-64" />
        </div>
        <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
          <div className="lg:col-span-7">
            <Skeleton className="h-130 w-full rounded-xl" />
          </div>
          <div className="lg:col-span-5">
            <Skeleton className="h-90 w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace="organizer"
        breadcrumbs={[
          { label: "My tournaments", href: "/organizer" },
          { label: "Host a tournament" },
        ]}
        title="Host a tournament"
      />

      {/* Main Form & Preview Workspace */}
      <div className="grid grid-cols-1 items-start gap-8 lg:grid-cols-12">
        {/* Form Column */}
        <div className="lg:col-span-7 xl:col-span-8">
          <HostTournamentForm
            organizerName={profile?.organizerName}
            onValuesChange={setDraft}
          />
        </div>

        {/* Sticky Preview Column */}
        <div className="lg:sticky lg:top-20 lg:col-span-5 xl:col-span-4">
          <HostTournamentPreview
            name={draft.name}
            game={draft.game}
            location={draft.location}
            description={draft.description}
            organizerName={profile?.organizerName}
          />
        </div>
      </div>
    </div>
  );
}
