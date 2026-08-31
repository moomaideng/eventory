import React from "react";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
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
          <Link
            href="/tournaments"
            className={cn(
              buttonVariants({ size: "lg" }),
              "flex w-full cursor-pointer items-center gap-2 shadow-xs sm:w-auto"
            )}
          >
            <Gamepad2 className="size-4" />
            <span>Explore Tournaments</span>
            <ArrowRight className="ml-1 size-4" />
          </Link>

          <Link
            href="/organizer/tournaments/new"
            className={cn(
              buttonVariants({ variant: "outline", size: "lg" }),
              "flex w-full cursor-pointer items-center gap-2 sm:w-auto"
            )}
          >
            <Trophy className="size-4" />
            <span>Host a Tournament</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
