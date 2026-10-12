import { Crown, Shield, ShieldCheck } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { formatRoleTitle } from "@/features/tournament-staff/utils";

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
        <Badge variant="default" className={className}>
          <Crown data-icon="inline-start" className="size-3" />
          Owner
        </Badge>
      );
    case "ADMIN":
      return (
        <Badge variant="secondary" className={className}>
          <ShieldCheck data-icon="inline-start" className="size-3" />
          Admin
        </Badge>
      );
    case "REFEREE":
      return (
        <Badge variant="outline" className={className}>
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
