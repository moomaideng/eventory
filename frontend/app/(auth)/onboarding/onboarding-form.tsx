"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useRole } from "@/context/role-context";
import { onboardUser } from "./actions";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { ArrowRight, Loader2, LogOut } from "lucide-react";

interface OnboardingFormProps {
  email?: string;
  avatarUrl?: string;
}

export function OnboardingForm({ email, avatarUrl }: OnboardingFormProps) {
  const router = useRouter();
  const { refreshUser, logout } = useRole();
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

  const handleSignOut = async () => {
    setIsSubmitting(true);
    try {
      await logout();
      router.push("/");
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
            {/* Authenticated Google Account Badge */}
            {email && (
              <div className="bg-muted/50 border-border/60 flex items-center justify-between gap-3 rounded-lg border p-2.5 text-left text-xs">
                <div className="flex items-center gap-2.5 overflow-hidden">
                  <Avatar className="size-6 shrink-0">
                    <AvatarImage src={avatarUrl} alt={email} />
                    <AvatarFallback className="text-[10px] uppercase">
                      {email.slice(0, 2)}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex flex-col truncate">
                    <span className="text-muted-foreground text-[10px]">
                      Signed in with Google
                    </span>
                    <span className="text-foreground truncate font-medium">
                      {email}
                    </span>
                  </div>
                </div>
              </div>
            )}

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

          <CardFooter className="flex flex-col gap-2 pt-2">
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

            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={handleSignOut}
              disabled={isSubmitting}
              className="text-muted-foreground hover:text-foreground flex items-center gap-1.5 text-xs"
            >
              <LogOut className="size-3.5" />
              <span>Sign out / Use a different account</span>
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
