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

interface OrganizerProfileViewProps {
  organizerName: string;
  organizerEmail?: string | null;
  onEdit: () => void;
}

export function OrganizerProfileView({
  organizerName,
  organizerEmail,
  onEdit,
}: OrganizerProfileViewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Organizer profile</CardTitle>
        <CardDescription>
          This information is shown to sponsors and participants across your
          tournaments.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Organizer name</p>
          <p className="font-medium">{organizerName}</p>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Contact email</p>
          <p className="font-medium">
            {organizerEmail || (
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
