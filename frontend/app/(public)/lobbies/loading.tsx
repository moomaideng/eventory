import { Skeleton } from "@/components/ui/skeleton";

export default function LobbiesLoading() {
  return (
    <div className="container mx-auto flex w-full max-w-md flex-1 items-center justify-center px-4 py-12 sm:px-8">
      <div className="border-border/60 bg-card flex w-full max-w-md flex-col gap-6 rounded-xl border p-6 shadow-xs">
        <div className="flex items-center gap-3">
          <Skeleton className="size-10 rounded-lg" />
          <div className="flex flex-col gap-1.5">
            <Skeleton className="h-6 w-44" />
            <Skeleton className="h-4 w-64" />
          </div>
        </div>
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-20" />
          <Skeleton className="h-10 w-full rounded-md" />
          <Skeleton className="h-3 w-48" />
        </div>
        <div className="flex justify-end">
          <Skeleton className="h-9 w-28 rounded-md" />
        </div>
      </div>
    </div>
  );
}
