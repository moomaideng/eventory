"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useRole } from "@/context/role-context";
import { onboardUser } from "./actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ArrowRight, Loader2 } from "lucide-react";

export function OnboardingForm() {
  const router = useRouter();
  const { refreshUser } = useRole();
  const [displayName, setDisplayName] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // React 19 Native Form Action
  const handleSubmit = async () => {
    const username = displayName.trim();
    if (!username) return;

    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      const result = await onboardUser(username);

      if (result.error) {
        setErrorMsg(result.error);
        return;
      }

      if (result.success) {
        await refreshUser();
        router.push("/");
      }
    } catch {
      setErrorMsg("An unexpected error occurred. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex flex-1 items-center justify-center px-4 py-16 sm:px-8">
      <Card className="w-full max-w-sm shadow-sm">
        <CardHeader className="text-center">
          <div className="bg-primary text-primary-foreground mx-auto flex size-10 items-center justify-center rounded-xl text-xl font-black">
            E
          </div>
          <CardTitle className="text-xl font-bold">
            Welcome to Eventory!
          </CardTitle>
          <CardDescription>
            Choose how you want your name to appear across tournaments and
            lobbies.
          </CardDescription>
        </CardHeader>

        <form action={handleSubmit}>
          <CardContent className="flex flex-col gap-4">
            {errorMsg && (
              <div className="bg-destructive/10 text-destructive border-destructive/20 rounded-lg border p-3 text-xs">
                {errorMsg}
              </div>
            )}

            <div className="flex flex-col gap-2 text-left">
              <label
                htmlFor="displayName"
                className="text-foreground text-xs font-medium"
              >
                Display Name
              </label>
              <Input
                id="displayName"
                name="displayName"
                type="text"
                placeholder="e.g. MooMai, ShadowNinja"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                required
                autoFocus
                disabled={isSubmitting}
                maxLength={32}
              />
            </div>
          </CardContent>

          <CardFooter className="pt-2">
            <Button
              type="submit"
              size="lg"
              disabled={!displayName.trim() || isSubmitting}
              className="flex w-full cursor-pointer items-center justify-center gap-2"
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="size-4 animate-spin" />
                  <span>Setting up account...</span>
                </>
              ) : (
                <>
                  <span>Get Started</span>
                  <ArrowRight className="size-4" />
                </>
              )}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
