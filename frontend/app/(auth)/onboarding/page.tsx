"use client";

import React, { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useRole } from "@/context/role-context";
import { createClient } from "@/lib/client";
import { apiClient } from "@/lib/api/client";
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

export default function OnboardingPage() {
  const router = useRouter();
  const { user, isLoading, refreshUser } = useRole();
  const [displayName, setDisplayName] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // If user already has an account, redirect straight to Home
  useEffect(() => {
    if (!isLoading && user) {
      router.replace("/");
    }
  }, [user, isLoading, router]);

  // React 19 Native Form Action Handler
  const handleSubmit = async () => {
    const username = displayName.trim();
    if (!username) return;

    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      // 1. Retrieve session access token
      const supabase = createClient();
      const {
        data: { session },
      } = await supabase.auth.getSession();

      const token = session?.access_token || "dev-token";

      // 2. Call Go Backend Onboarding API directly via openapi-fetch
      const { data, error } = await apiClient.POST(
        "/api/v1/accounts/onboard",
        {
          body: {
            username,
          },
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (error) {
        setErrorMsg(
          error.detail || "Failed to create account. Please try another username."
        );
        setIsSubmitting(false);
        return;
      }

      if (data) {
        // 3. Sync RoleContext with newly created account
        await refreshUser();
        router.push("/");
      }
    } catch (err) {
      setErrorMsg("Unable to connect to server. Please try again.");
      console.error("Onboarding error:", err);
      setIsSubmitting(false);
    }
  };

  // Prevent form flashing while auth state is hydrating
  if (isLoading || user) {
    return (
      <div className="flex flex-1 items-center justify-center px-4 py-16 sm:px-8">
        <Card className="flex w-full max-w-sm items-center justify-center py-12 shadow-sm">
          <Loader2 className="text-primary size-6 animate-spin" />
        </Card>
      </div>
    );
  }

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
              <div className="bg-destructive/10 text-destructive rounded-lg border border-destructive/20 p-3 text-xs">
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
