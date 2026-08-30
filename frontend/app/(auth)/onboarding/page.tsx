"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import { useRole } from "@/context/role-context";
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
import { ArrowRight } from "lucide-react";

export default function OnboardingPage() {
  const router = useRouter();
  const { user, updateUserProfile } = useRole();
  const [displayName, setDisplayName] = useState(user?.displayName || "");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) return;

    updateUserProfile(displayName.trim());
    router.push("/");
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

        <form onSubmit={handleSubmit}>
          <CardContent className="flex flex-col gap-4">
            <div className="flex flex-col gap-2 text-left">
              <label
                htmlFor="displayName"
                className="text-foreground text-xs font-medium"
              >
                Display Name
              </label>
              <Input
                id="displayName"
                type="text"
                placeholder="e.g. MooMai, ShadowNinja"
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                required
                autoFocus
              />
            </div>
          </CardContent>

          <CardFooter className="pt-2">
            <Button
              type="submit"
              size="lg"
              disabled={!displayName.trim()}
              className="flex w-full cursor-pointer items-center justify-center gap-2"
            >
              <span>Get Started</span>
              <ArrowRight className="size-4" />
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}
