"use client";

import { useEffect } from "react";
import { Button } from "@/components/ui/button";
import { AlertCircle, RotateCcw } from "lucide-react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("App boundary caught error:", error);
  }, [error]);

  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 py-16 text-center sm:px-8">
      <div className="bg-destructive/10 text-destructive mb-4 flex size-12 items-center justify-center rounded-xl">
        <AlertCircle className="size-6" />
      </div>
      <h2 className="text-foreground text-2xl font-bold tracking-tight">
        Something went wrong
      </h2>
      <p className="text-muted-foreground mt-2 max-w-md text-sm">
        An unexpected error occurred while loading this page. You can try
        refreshing or reloading the action.
      </p>
      <div className="mt-6 flex items-center gap-3">
        <Button variant="outline" onClick={() => reset()}>
          <RotateCcw data-icon="inline-start" />
          Try Again
        </Button>
      </div>
    </div>
  );
}
