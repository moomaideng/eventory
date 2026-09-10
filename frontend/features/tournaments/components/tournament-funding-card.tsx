"use client";

import React from "react";
import { Trophy } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import type { components } from "@/lib/api/schema";
import { formatMoney, formatPercentage } from "../utils";

type Funding = components["schemas"]["TournamentFundingResponse"];

const numberFormatter = new Intl.NumberFormat("en-US");

export function TournamentFundingCard({ funding }: { funding: Funding }) {
  const progressValue = Math.min(Math.max(funding.percentage, 0), 100);

  return (
    <Card>
      <CardHeader>
        <div className="bg-primary text-primary-foreground flex size-10 items-center justify-center rounded-lg">
          <Trophy className="size-5" />
        </div>
        <CardTitle>Funding progress</CardTitle>
        <CardDescription>
          Community support helps fund the prize pool and event costs.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-5">
        <div className="flex flex-col gap-1">
          <p className="text-3xl font-bold tracking-tight">
            {formatMoney(funding.raisedAmount, funding.currency)}
          </p>
          <p className="text-muted-foreground text-sm">
            raised of {formatMoney(funding.goalAmount, funding.currency)}
          </p>
        </div>
        <Progress value={progressValue}>
          <ProgressLabel>Funding goal</ProgressLabel>
          <ProgressValue>
            {() => formatPercentage(funding.percentage)}
          </ProgressValue>
        </Progress>
        <Separator />
        <div className="grid grid-cols-2 gap-4">
          <div className="flex flex-col gap-1">
            <p className="text-muted-foreground text-xs">Supporters</p>
            <p className="font-medium">
              {numberFormatter.format(funding.supporterCount)}
            </p>
          </div>
          <div className="flex flex-col gap-1">
            <p className="text-muted-foreground text-xs">Still needed</p>
            <p className="font-medium">
              {formatMoney(funding.remainingAmount, funding.currency)}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
