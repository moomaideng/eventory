"use client";

import { useState } from "react";
import { X } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableCaption,
} from "@/components/ui/table";
import { formatDate, initials } from "@/features/tournaments/utils";
import { StaffRoleBadge } from "@/features/tournament-staff/components/staff-role-badge";
import { UpdateStaffRoleDialog } from "@/features/tournament-staff/components/update-staff-role-dialog";
import { RemoveStaffDialog } from "@/features/tournament-staff/components/remove-staff-dialog";
import type { StaffMember } from "@/features/tournament-staff/utils";

export function StaffTable({
  staff,
  tournamentId,
  isOwner,
}: {
  staff: StaffMember[];
  tournamentId: string;
  isOwner: boolean;
}) {
  const [editingStaff, setEditingStaff] = useState<StaffMember | null>(null);
  const [removingStaff, setRemovingStaff] = useState<StaffMember | null>(null);

  if (!staff.length) {
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyTitle>No staff members</EmptyTitle>
          <EmptyDescription>
            Invite admins or referees to help manage this tournament.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <>
      <div className="overflow-hidden rounded-xl border border-border/40">
        <Table className="table-fixed">
          <TableCaption className="sr-only">
            Tournament staff roster and assigned roles
          </TableCaption>
          <TableHeader>
            <TableRow>
              <TableHead className="w-1/2 sm:w-2/5">Member</TableHead>
              <TableHead className="w-28 sm:w-36">Role</TableHead>
              <TableHead className="hidden w-36 sm:table-cell">Joined</TableHead>
              {isOwner ? (
                <TableHead className="w-40 text-right sm:w-44">Actions</TableHead>
              ) : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {staff.map((member) => {
              const isMemberOwner = member.role === "OWNER";
              return (
                <TableRow key={member.accountId}>
                  <TableCell className="align-middle">
                    <div className="flex items-center gap-3 py-1">
                      <Avatar className="size-8 shrink-0">
                        <AvatarFallback>
                          {initials(member.displayName || member.handle)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex min-w-0 flex-col">
                        <span className="font-medium truncate">
                          {member.displayName || member.handle}
                        </span>
                        <span className="text-muted-foreground text-xs truncate">
                          @{member.handle}
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="align-middle">
                    <StaffRoleBadge role={member.role} />
                  </TableCell>
                  <TableCell className="text-muted-foreground hidden align-middle text-sm sm:table-cell">
                    {formatDate(member.createdAt)}
                  </TableCell>
                  {isOwner ? (
                    <TableCell className="align-middle text-right">
                      {isMemberOwner ? (
                        <span className="text-muted-foreground text-xs italic">
                          Protected
                        </span>
                      ) : (
                        <div className="flex items-center justify-end gap-1.5">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setEditingStaff(member)}
                          >
                            Change role
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon-sm"
                            className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                            aria-label={`Remove @${member.handle}`}
                            onClick={() => setRemovingStaff(member)}
                          >
                            <X className="size-4" />
                          </Button>
                        </div>
                      )}
                    </TableCell>
                  ) : null}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>

      <UpdateStaffRoleDialog
        tournamentId={tournamentId}
        staff={editingStaff}
        open={Boolean(editingStaff)}
        onOpenChange={(open) => {
          if (!open) setEditingStaff(null);
        }}
      />

      <RemoveStaffDialog
        tournamentId={tournamentId}
        staff={removingStaff}
        open={Boolean(removingStaff)}
        onOpenChange={(open) => {
          if (!open) setRemovingStaff(null);
        }}
      />
    </>
  );
}
