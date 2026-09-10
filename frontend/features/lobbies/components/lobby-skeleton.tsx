import React from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function LobbyWorkspaceSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <Skeleton className="h-5 w-32" />
      <Skeleton className="h-12 w-72" />
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Skeleton className="h-96 w-full" />
        <div className="flex flex-col gap-6">
          <Skeleton className="h-64 w-full" />
          <Skeleton className="h-56 w-full" />
        </div>
      </div>
    </div>
  );
}
