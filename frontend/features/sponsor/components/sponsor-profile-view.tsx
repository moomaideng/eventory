"use client";

import { Pencil } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface SponsorProfileViewProps {
  sponsorName: string;
  sponsorEmail?: string | null;
  onEdit: () => void;
}

export function SponsorProfileView({
  sponsorName,
  sponsorEmail,
  onEdit,
}: SponsorProfileViewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Sponsor profile</CardTitle>
        <CardDescription>
          This information is shown to organizers when you sponsor a
          tournament.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Sponsor name</p>
          <p className="font-medium">{sponsorName}</p>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Contact email</p>
          <p className="font-medium">
            {sponsorEmail || (
              <span className="text-muted-foreground font-normal">
                Not set
              </span>
            )}
          </p>
        </div>
      </CardContent>
      <CardFooter className="justify-end">
        <Button onClick={onEdit}>
          <Pencil data-icon="inline-start" />
          Edit profile
        </Button>
      </CardFooter>
    </Card>
  );
}
