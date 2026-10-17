"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";

interface HostTournamentFundingCardProps {
  enabled: boolean;
  onEnabledChange: (value: boolean) => void;
  goal: number;
  onGoalChange: (value: number) => void;
  deadline: string;
  onDeadlineChange: (value: string) => void;
  errors: Partial<Record<"fundingGoal" | "fundingDeadline", string>>;
}

export function HostTournamentFundingCard({
  enabled,
  onEnabledChange,
  goal,
  onGoalChange,
  deadline,
  onDeadlineChange,
  errors,
}: HostTournamentFundingCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Crowdfunding</CardTitle>
        <CardDescription>
          Start with sponsorship and funding first, or open participant
          registration now.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field>
            <FieldLabel>Launch stage</FieldLabel>
            <RadioGroup
              value={enabled ? "crowdfunding" : "registration"}
              onValueChange={(value) =>
                onEnabledChange(value === "crowdfunding")
              }
              className="flex flex-col gap-3 sm:flex-row"
            >
              <FieldLabel
                htmlFor="stage-registration"
                className="flex flex-1 cursor-pointer items-center justify-between rounded-lg border p-3"
              >
                <span>Open registration now</span>
                <RadioGroupItem id="stage-registration" value="registration" />
              </FieldLabel>
              <FieldLabel
                htmlFor="stage-crowdfunding"
                className="flex flex-1 cursor-pointer items-center justify-between rounded-lg border p-3"
              >
                <span>Fund the event first</span>
                <RadioGroupItem id="stage-crowdfunding" value="crowdfunding" />
              </FieldLabel>
            </RadioGroup>
          </Field>
          {enabled ? (
            <>
              <Field data-invalid={Boolean(errors.fundingGoal)}>
                <FieldLabel htmlFor="funding-goal">
                  Funding goal (THB)
                </FieldLabel>
                <Input
                  id="funding-goal"
                  type="number"
                  min={1}
                  step={1}
                  value={goal || ""}
                  onChange={(event) => onGoalChange(Number(event.target.value))}
                  aria-invalid={Boolean(errors.fundingGoal)}
                />
                {errors.fundingGoal ? (
                  <FieldError>{errors.fundingGoal}</FieldError>
                ) : (
                  <FieldDescription>
                    The amount you want to raise for this event.
                  </FieldDescription>
                )}
              </Field>
              <Field data-invalid={Boolean(errors.fundingDeadline)}>
                <FieldLabel htmlFor="funding-deadline">
                  Funding deadline
                </FieldLabel>
                <Input
                  id="funding-deadline"
                  type="datetime-local"
                  value={deadline}
                  onChange={(event) => onDeadlineChange(event.target.value)}
                  aria-invalid={Boolean(errors.fundingDeadline)}
                />
                {errors.fundingDeadline ? (
                  <FieldError>{errors.fundingDeadline}</FieldError>
                ) : (
                  <FieldDescription>
                    Must be in the future and before the registration deadline.
                  </FieldDescription>
                )}
              </Field>
              <p className="text-muted-foreground text-sm">
                People can view the fundraising campaign while participant
                registration stays closed.
              </p>
            </>
          ) : null}
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
