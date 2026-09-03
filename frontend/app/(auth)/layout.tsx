import React from "react";
import Link from "next/link";
import { ThemeToggle } from "@/components/theme-toggle";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-full flex-col">
      {/* Minimal Header with Brand Logo & Theme Toggle */}
      <header className="bg-background w-full border-b">
        <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
          <Link href="/" className="flex items-center gap-2">
            <div className="bg-primary text-primary-foreground flex size-9 items-center justify-center rounded-lg font-black">
              E
            </div>
            <span className="text-foreground text-xl font-bold">
              Eventory<span className="text-primary">.</span>
            </span>
          </Link>

          <ThemeToggle />
        </div>
      </header>

      <main className="flex flex-1 flex-col">{children}</main>
    </div>
  );
}
