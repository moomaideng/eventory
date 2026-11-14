import type { UserProfile } from "./auth-context";

export type DevProfileId = "alex-dev" | "bob-dev" | "carter-dev";

export interface DevProfileOption {
  id: DevProfileId;
  token: DevProfileId;
  roleLabel: string;
  description: string;
  user: UserProfile;
}

export const DEV_PROFILES: Record<DevProfileId, DevProfileOption> = {
  "alex-dev": {
    id: "alex-dev",
    token: "alex-dev",
    roleLabel: "All-rounder",
    description: "Alex Events · Alex Ventures",
    user: {
      id: "99999999-0000-4000-8000-000000000001",
      email: "alex.dev@eventory.gg",
      displayName: "Alex Dev",
      handle: "alex_dev",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=AlexDev",
      status: "ACTIVE",
    },
  },
  "bob-dev": {
    id: "bob-dev",
    token: "bob-dev",
    roleLabel: "Organizer",
    description: "Bob Events (no Sponsor)",
    user: {
      id: "99999999-0000-4000-8000-000000000002",
      email: "bob.dev@eventory.gg",
      displayName: "Bob Dev",
      handle: "bob_dev",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=BobDev",
      status: "ACTIVE",
    },
  },
  "carter-dev": {
    id: "carter-dev",
    token: "carter-dev",
    roleLabel: "Sponsor",
    description: "Carter Fundings (no Organizer)",
    user: {
      id: "99999999-0000-4000-8000-000000000003",
      email: "carter.dev@eventory.gg",
      displayName: "Carter Dev",
      handle: "carter_dev",
      avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=CarterDev",
      status: "ACTIVE",
    },
  },
};

export const DEV_PROFILE_LIST: DevProfileOption[] = [
  DEV_PROFILES["alex-dev"],
  DEV_PROFILES["bob-dev"],
  DEV_PROFILES["carter-dev"],
];

export function parseDevSessionCookie(
  value: string | undefined
): DevProfileId | null {
  if (!value) return null;
  if (value === "true" || value === "alex-dev") return "alex-dev";
  if (value === "bob-dev" || value === "carter-dev") return value;
  return null;
}

// Default primary mock user profile for local development mode (Alex Dev)
export const MOCK_USER: UserProfile = DEV_PROFILES["alex-dev"].user;
