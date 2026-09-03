import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Home } from "lucide-react";

export default function NotFound() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center px-4 py-20 text-center sm:px-8">
      <p className="text-primary font-mono text-sm font-semibold tracking-wider uppercase">
        404 Error
      </p>
      <h1 className="text-foreground mt-2 text-4xl font-extrabold tracking-tight sm:text-5xl">
        Page Not Found
      </h1>
      <p className="text-muted-foreground mt-3 max-w-md text-base">
        The tournament, lobby, or page you are looking for does not exist or has
        been moved.
      </p>
      <div className="mt-8">
        <Button size="lg" render={<Link href="/" />} nativeButton={false}>
          <Home data-icon="inline-start" />
          Back to Home
        </Button>
      </div>
    </div>
  );
}
