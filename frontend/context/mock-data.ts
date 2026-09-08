import type { UserProfile } from "./auth-context";
import type { OrganizerProfile, SponsorProfile } from "./role-context";

// Mock primary user profile for local development mode
export const MOCK_USER: UserProfile = {
  id: "99999999-0000-0000-0000-000000000001",
  email: "dev@eventory.gg",
  displayName: "Dev Competitor",
  handle: "dev_competitor",
  avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=MooMai",
};

// Default organizer profile for in-memory workspace testing
export const DEFAULT_ORGANIZER: OrganizerProfile = {
  id: "org-1",
  name: "Chula Esports Club",
  bio: "Official university esports club hosting regional tournaments.",
};

// Default sponsor profile for in-memory workspace testing
export const DEFAULT_SPONSOR: SponsorProfile = {
  id: "sp-1",
  companyName: "Red Bull Gaming",
  websiteUrl: "https://redbull.com",
};
