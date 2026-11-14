import React from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function TeamCreateSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-3 w-52" />
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-9 w-64" />
        <Skeleton className="h-5 w-80" />
      </div>
      <Skeleton className="h-52 w-full rounded-xl" />
      <Skeleton className="h-48 w-full rounded-xl" />
    </div>
  );
}
