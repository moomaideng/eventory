import { Skeleton } from "@/components/ui/skeleton";

export default function AuthLoading() {
  return (
    <div className="flex flex-1 items-center justify-center px-4 py-16 sm:px-8">
      <div className="border-border/60 bg-card flex w-full max-w-sm flex-col gap-6 rounded-xl border p-6 shadow-xs">
        <div className="flex flex-col items-center gap-3">
          <Skeleton className="size-10 rounded-xl" />
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-60" />
        </div>
        <div className="flex flex-col gap-4">
          <Skeleton className="h-10 w-full rounded-md" />
          <Skeleton className="h-10 w-full rounded-md" />
        </div>
      </div>
    </div>
  );
}
