import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { RegistrationFormView } from "@/features/registration-form/components/registration-form-view";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Registration Form - Eventory",
  description: "Fill in the tournament registration form.",
};

export default async function TournamentFormPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <RegistrationFormView tournamentId={parsed.data} />;
}
