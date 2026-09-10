"use client";

import React from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LogIn } from "lucide-react";

export function LobbySignInCard({ onDevLogin }: { onDevLogin: () => void }) {
  return (
    <div className="container mx-auto flex w-full max-w-3xl flex-1 items-center px-4 py-12 sm:px-8">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>Sign in to open this invitation</CardTitle>
          <CardDescription>
            You need an Eventory account before you can view or join a team roster.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-sm">
            Sign in with your verified identity or use quick access in development mode.
          </p>
        </CardContent>
        <CardFooter className="flex flex-wrap gap-3">
          <Button
            render={<Link href="/login" />}
            nativeButton={false}
          >
            <LogIn data-icon="inline-start" />
            Sign in
          </Button>
          <Button variant="outline" onClick={onDevLogin}>
            Use Dev Quick Login
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
