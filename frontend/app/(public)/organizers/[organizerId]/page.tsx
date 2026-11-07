import React, { Suspense } from "react";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { OrganizerPublicView } from "@/features/organizer/components/organizer-public-view";
import { OrganizerDetailsSkeleton } from "@/features/organizer/components/organizer-details-skeleton";
import { organizerIdParamSchema } from "@/features/organizer/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ organizerId: string }>;
}): Promise<Metadata> {
  const { organizerId } = await params;
  return {
    title: "Organizer Profile - Eventory",
    description: `View organizer profile and all tournaments hosted by organizer ${organizerId}.`,
  };
}

export default async function OrganizerPage({
  params,
}: {
  params: Promise<{ organizerId: string }>;
}) {
  const { organizerId } = await params;
  const parsed = organizerIdParamSchema.safeParse(organizerId);
  if (!parsed.success) {
    notFound();
  }

  return (
    <Suspense fallback={<OrganizerDetailsSkeleton />}>
      <OrganizerPublicView organizerId={parsed.data} />
    </Suspense>
  );
}
