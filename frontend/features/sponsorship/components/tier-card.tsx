import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import type { SponsorshipTier } from "@/features/sponsorship/utils";
import { formatMoney } from "@/features/tournaments/utils";

export function TierCard({
  tier,
  currency,
  footer,
}: {
  tier: SponsorshipTier;
  currency: string;
  footer?: ReactNode;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="wrap-anywhere">{tier.name}</CardTitle>
        <CardDescription>{formatMoney(tier.price, currency)}</CardDescription>
        <CardAction>
          <Badge variant="secondary">
            {tier.slots} {tier.slots === 1 ? "slot" : "slots"}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent>
        <ul className="text-muted-foreground flex list-disc flex-col gap-1 pl-5 text-sm">
          {(tier.perks ?? []).map((perk) => (
            <li key={perk} className="wrap-anywhere">
              {perk}
            </li>
          ))}
        </ul>
      </CardContent>
      {footer ? <CardFooter className="gap-2">{footer}</CardFooter> : null}
    </Card>
  );
}
