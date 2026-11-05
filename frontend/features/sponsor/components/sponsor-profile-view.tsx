"use client";

import { Badge } from "@/components/ui/badge";
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
  tags: string[];
  onEdit: () => void;
}

export function SponsorProfileView({
  sponsorName,
  sponsorEmail,
  tags,
  onEdit,
}: SponsorProfileViewProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Sponsor profile</CardTitle>
        <CardDescription>
          Your sponsor name, contact email and tags are visible to organizers in
          the sponsor directory.
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
              <span className="text-muted-foreground font-normal">Not set</span>
            )}
          </p>
        </div>
        <div className="border-border space-y-2 border-t pt-4">
          <p className="font-medium">Sponsor directory</p>
          <p className="text-muted-foreground text-sm">Visible to organizers</p>
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <Badge key={tag} variant="secondary">
                {tag}
              </Badge>
            ))}
          </div>
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
