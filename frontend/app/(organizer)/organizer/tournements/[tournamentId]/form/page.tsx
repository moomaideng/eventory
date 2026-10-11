import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { RegistrationFormBuilder } from "@/features/organizer/components/registration-form-builder";
import { tournamentIdParamSchema } from "@/features/tournaments/schemas";

export const metadata: Metadata = {
  title: "Registration Form - Eventory",
  description: "Configure tournament registration form.",
};

export default async function RegistrationFormPage({
  params,
}: {
  params: Promise<{ tournamentId: string }>;
}) {
  const { tournamentId } = await params;
  const parsed = tournamentIdParamSchema.safeParse(tournamentId);
  if (!parsed.success) {
    notFound();
  }

  return <RegistrationFormBuilder tournamentId={parsed.data} />;
}
