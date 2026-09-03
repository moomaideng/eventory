import React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRight, Trophy, Gamepad2 } from "lucide-react";

export default function HomePage() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 py-20 sm:px-8">
      <div className="flex w-full max-w-3xl flex-col gap-8 text-center">
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

        {/* Clean Primary Actions */}
        <div className="flex flex-col items-center justify-center gap-3 pt-2 sm:flex-row">
          <Button
            size="lg"
            render={<Link href="/tournaments" />}
            nativeButton={false}
            className="w-full sm:w-auto"
          >
            <Gamepad2 data-icon="inline-start" />
            Explore Tournaments
            <ArrowRight data-icon="inline-end" />
          </Button>

          <Button
            variant="outline"
            size="lg"
            render={<Link href="/organizer/tournaments/new" />}
            nativeButton={false}
            className="w-full sm:w-auto"
          >
            <Trophy data-icon="inline-start" />
            Host a Tournament
          </Button>
        </div>
      </div>
    </div>
  );
}
