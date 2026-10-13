import { Crown, Shield, ShieldCheck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { formatRoleTitle } from "@/features/tournament-staff/utils";
import { cn } from "@/lib/utils";

export function StaffRoleBadge({
  role,
  className,
}: {
  role: string;
  className?: string;
}) {
  switch (role) {
    case "OWNER":
      return (
        <Badge
          variant="outline"
          className={cn(
            "border-amber-500/30 bg-amber-500/15 text-amber-600 dark:text-amber-400 font-medium",
            className
          )}
        >
          <Crown data-icon="inline-start" className="size-3" />
          Owner
        </Badge>
      );
    case "ADMIN":
      return (
        <Badge
          variant="outline"
          className={cn(
            "border-blue-500/30 bg-blue-500/15 text-blue-600 dark:text-blue-400 font-medium",
            className
          )}
        >
          <ShieldCheck data-icon="inline-start" className="size-3" />
          Admin
        </Badge>
      );
    case "REFEREE":
      return (
        <Badge
          variant="outline"
          className={cn(
            "border-purple-500/30 bg-purple-500/15 text-purple-600 dark:text-purple-400 font-medium",
            className
          )}
        >
          <Shield data-icon="inline-start" className="size-3" />
          Referee
        </Badge>
      );
    default:
      return (
        <Badge variant="outline" className={className}>
          {formatRoleTitle(role)}
        </Badge>
      );
  }
}
