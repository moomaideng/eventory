import { z } from "zod";

export const inviteStaffSchema = z.object({
  handle: z
    .string({ message: "Handle is required." })
    .trim()
    .min(3, "Handle must be at least 3 characters.")
    .max(32, "Keep handle under 32 characters.")
    .regex(
      /^[a-zA-Z0-9_.-]+$/,
      "Handle may only contain letters, numbers, underscores, dots, and hyphens."
    ),
  role: z.enum(["ADMIN", "REFEREE"], {
    message: "Select a valid staff role.",
  }),
});

export type InviteStaffInput = z.infer<typeof inviteStaffSchema>;

export const updateStaffRoleSchema = z.object({
  role: z.enum(["ADMIN", "REFEREE"], {
    message: "Select a valid staff role.",
  }),
});

export type UpdateStaffRoleInput = z.infer<typeof updateStaffRoleSchema>;
