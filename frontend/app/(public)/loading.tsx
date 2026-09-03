import { Skeleton } from "@/components/ui/skeleton";

export default function PublicLoading() {
  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-20 sm:px-8">
      {/* Hero Header Skeleton */}
      <div className="flex flex-col items-center gap-4 text-center">
        <Skeleton className="h-12 w-3/4 max-w-lg" />
        <Skeleton className="h-5 w-1/2 max-w-md" />
        <div className="mt-2 flex gap-3">
          <Skeleton className="h-10 w-44 rounded-md" />
          <Skeleton className="h-10 w-44 rounded-md" />
        </div>
      </div>
    </div>
  );
}
