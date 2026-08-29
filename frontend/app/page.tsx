'use client';

import React from 'react';
import Link from 'next/link';
import { buttonVariants } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { ArrowRight, Trophy, Gamepad2 } from 'lucide-react';

export default function HomePage() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center px-4 py-20 sm:px-8">
      <div className="max-w-3xl w-full text-center space-y-8">
        {/* Clean Hero Title & Description */}
        <div className="space-y-4">
          <h1 className="text-4xl font-extrabold tracking-tight sm:text-6xl text-foreground">
            Host, Compete, and Sponsor Tournaments.
          </h1>
          <p className="max-w-xl mx-auto text-base sm:text-lg text-muted-foreground">
            A single identity for competitors, organizers, and sponsors. Manage brackets,
            crowdfund prize pools, and organize team lobbies effortlessly.
          </p>
        </div>

        {/* Clean Primary Actions */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <Link
            href="/tournaments"
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full sm:w-auto flex items-center gap-2 cursor-pointer shadow-xs'
            )}
          >
            <Gamepad2 className="h-4 w-4" />
            <span>Explore Tournaments</span>
            <ArrowRight className="h-4 w-4 ml-1" />
          </Link>

          <Link
            href="/organizer/tournaments/new"
            className={cn(
              buttonVariants({ variant: 'outline', size: 'lg' }),
              'w-full sm:w-auto flex items-center gap-2 cursor-pointer'
            )}
          >
            <Trophy className="h-4 w-4" />
            <span>Host a Tournament</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
