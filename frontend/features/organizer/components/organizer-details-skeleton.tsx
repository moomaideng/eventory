import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";

export function OrganizerDetailsSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Skeleton className="h-9 w-36" />
      <Card className="overflow-hidden border py-0">
        <div className="grid grid-cols-1 items-stretch gap-6 p-6 sm:p-8 md:grid-cols-[1fr_auto_1fr] md:gap-8">
          <div className="flex items-center gap-4">
            <Skeleton className="size-16 rounded-xl" />
            <div className="flex flex-col gap-2">
              <div className="flex items-center gap-2">
                <Skeleton className="h-8 w-48 sm:w-64" />
                <Skeleton className="h-5 w-20 rounded-full" />
              </div>
              <Skeleton className="h-4 w-56" />
            </div>
          </div>

          <Separator orientation="vertical" className="hidden self-stretch md:block" />
          <Separator orientation="horizontal" className="md:hidden" />

          <div className="flex flex-col justify-center gap-3">
            <Skeleton className="h-4 w-44" />
            <Skeleton className="h-4 w-52" />
            <Skeleton className="h-4 w-36" />
          </div>
        </div>
      </Card>

      <div className="flex flex-col gap-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-6 w-24" />
        </div>
        <TournamentGridSkeleton />
      </div>
    </div>
  );
}
