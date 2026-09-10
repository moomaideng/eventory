import type { UserProfile } from "./auth-context";

// Mock primary user profile for local development mode (Alex (Dev))
export const MOCK_USER: UserProfile = {
  id: "99999999-0000-4000-8000-000000000001",
  email: "dev@eventory.gg",
  displayName: "Alex (Dev)",
  handle: "dev_alex",
  avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=AlexDev",
};
