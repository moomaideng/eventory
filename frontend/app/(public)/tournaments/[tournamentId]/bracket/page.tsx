import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { PublicBracketViewer } from "@/features/bracket/components/public-bracket-viewer";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export async function generateMetadata({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}): Promise<Metadata> {
  const { tournamentId } = await params;
  return {
    title: `Tournament Bracket - Eventory`,
    description: `View tournament bracket progression and match results for tournament ${tournamentId}.`,
  };
}

export default async function PublicBracketPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-8 sm:px-8">
      <PublicBracketViewer tournamentId={parsed.data} />
    </div>
  );
}
