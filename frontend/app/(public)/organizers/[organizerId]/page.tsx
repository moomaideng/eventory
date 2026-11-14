import { Suspense } from "react";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { OrganizerProfileSkeleton } from "@/features/organizers/components/organizer-profile-skeleton";
import { OrganizerPublicProfile } from "@/features/organizers/components/organizer-public-profile";
import { organizerIdParamSchema } from "@/features/organizers/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ organizerId: string }>;
}): Promise<Metadata> {
  const { organizerId } = await params;
  return {
    title: `Organizer Profile - Eventory`,
    description: `View organizer profile details and hosted tournaments for organizer ${organizerId}.`,
  };
}

export default async function OrganizerProfilePage({
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
    <Suspense fallback={<OrganizerProfileSkeleton />}>
      <OrganizerPublicProfile organizerId={parsed.data} />
    </Suspense>
  );
}
