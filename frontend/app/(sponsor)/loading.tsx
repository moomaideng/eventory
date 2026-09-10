import { Skeleton } from "@/components/ui/skeleton";

export default function SponsorLoading() {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
      <div className="flex items-center gap-3">
        <Skeleton className="h-8 w-36 rounded-md" />
        <Skeleton className="h-8 w-40 rounded-md" />
      </div>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-2">
          <Skeleton className="h-4 w-32" />
          <Skeleton className="h-8 w-64" />
        </div>
        <Skeleton className="h-48 w-full rounded-xl" />
      </div>
    </div>
  );
}
