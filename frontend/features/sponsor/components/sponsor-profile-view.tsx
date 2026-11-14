"use client";

import Image from "next/image";
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
  phone?: string | null;
  websiteUrl?: string | null;
  logoUrl?: string | null;
  tags: string[];
  onEdit: () => void;
}

export function SponsorProfileView({
  sponsorName,
  sponsorEmail,
  phone,
  websiteUrl,
  logoUrl,
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
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Phone</p>
          <p className="font-medium">
            {phone || (
              <span className="text-muted-foreground font-normal">Not set</span>
            )}
          </p>
        </div>
        <div className="flex flex-col gap-1">
          <p className="text-muted-foreground text-sm">Brand website</p>
          <p className="font-medium break-all">
            {websiteUrl ? (
              <a
                href={websiteUrl}
                target="_blank"
                rel="noopener noreferrer"
                className="underline underline-offset-4"
              >
                {websiteUrl}
              </a>
            ) : (
              <span className="text-muted-foreground font-normal">Not set</span>
            )}
          </p>
        </div>
        <div className="flex flex-col gap-2">
          <p className="text-muted-foreground text-sm">Brand logo</p>
          {logoUrl ? (
            <Image
              src={logoUrl}
              alt={`${sponsorName} logo`}
              width={160}
              height={64}
              unoptimized
              className="bg-background h-16 w-40 rounded-md border object-contain p-2"
            />
          ) : (
            <p className="text-muted-foreground text-sm">Not set</p>
          )}
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
