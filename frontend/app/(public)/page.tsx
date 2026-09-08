import React from "react";
import type { Metadata } from "next";
import { createClient } from "@/lib/server";
import { HeroActions } from "./hero-actions";

export const metadata: Metadata = {
  title: "Eventory - Host, Compete, and Sponsor Tournaments",
  description:
    "A single unified identity for competitors, organizers, and sponsors. Manage brackets, crowdfund prize pools, and organize team lobbies effortlessly.",
};

export default async function HomePage() {
  let serverAuthenticated = false;

  try {
    const supabase = await createClient();
    const { data } = await supabase.auth.getClaims();
    serverAuthenticated = Boolean(data?.claims);
  } catch {
    // Graceful fallback if session check fails
  }

  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 py-24 sm:px-8">
      <div className="flex w-full max-w-3xl flex-col items-center gap-8 text-center">
        {/* Clean Hero Title & Description */}
        <div className="flex flex-col gap-4">
          <h1 className="text-foreground text-4xl font-extrabold tracking-tight sm:text-6xl">
            Host, Compete, and Sponsor Tournaments.
          </h1>
          <p className="text-muted-foreground mx-auto max-w-xl text-base sm:text-lg">
            A single identity for competitors, organizers, and sponsors. Manage
            brackets, crowdfund prize pools, and organize team lobbies
            effortlessly.
          </p>
        </div>

        {/* Clear, balanced primary action buttons */}
        <HeroActions serverAuthenticated={serverAuthenticated} />
      </div>
    </div>
  );
}
