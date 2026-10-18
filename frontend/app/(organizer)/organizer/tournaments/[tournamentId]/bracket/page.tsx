import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BracketWorkspace } from "@/features/bracket/components/bracket-workspace";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Tournament Bracket & Matches - Eventory",
  description: "Customize tournament bracket layout, seed teams, and record live match scores.",
};

export default async function BracketPage({
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
      <div>
        <Button
          variant="outline"
          size="sm"
          render={<Link href={`/organizer/tournaments/${parsed.data}`} />}
          nativeButton={false}
        >
          <ArrowLeft data-icon="inline-start" />
          Back to Workspace
        </Button>
      </div>

      <BracketWorkspace tournamentId={parsed.data} />
    </div>
  );
}
