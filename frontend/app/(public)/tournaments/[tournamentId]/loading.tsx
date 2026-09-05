import { Skeleton } from "@/components/ui/skeleton";

export default function TournamentDetailsLoading() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-10 sm:px-8">
      <Skeleton className="h-9 w-36" />
      <div className="flex flex-col gap-4">
        <Skeleton className="h-6 w-28" />
        <Skeleton className="h-12 w-full max-w-2xl" />
        <Skeleton className="h-5 w-full max-w-xl" />
      </div>
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          <Skeleton className="h-52 w-full" />
          <Skeleton className="h-80 w-full" />
        </div>
        <Skeleton className="h-96 w-full" />
      </div>
    </div>
  );
}
