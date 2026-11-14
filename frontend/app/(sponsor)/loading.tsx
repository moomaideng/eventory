import { Skeleton } from "@/components/ui/skeleton";

export default function SponsorLoading() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-9 w-64" />
      </div>
      <Skeleton className="h-48 w-full rounded-xl" />
    </div>
  );
}
