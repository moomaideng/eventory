"use client";

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useMemo,
} from "react";
import { useAuth } from "./auth-context";
import { DEFAULT_ORGANIZER, DEFAULT_SPONSOR } from "./mock-data";

export type UserRole = "competitor" | "organizer" | "sponsor";

export interface OrganizerProfile {
  id: string;
  name: string;
  bio?: string;
  logoUrl?: string;
}

export interface SponsorProfile {
  id: string;
  companyName: string;
  websiteUrl?: string;
  logoUrl?: string;
}

export interface RoleContextType {
  activeRole: UserRole;
  activeProfileName: string;
  organizerProfile: OrganizerProfile | null;
  sponsorProfile: SponsorProfile | null;
  setRole: (role: UserRole) => void;
  createOrUpdateOrganizerProfile: (name: string, bio?: string) => void;
  createOrUpdateSponsorProfile: (
    companyName: string,
    websiteUrl?: string
  ) => void;
}

const RoleContext = createContext<RoleContextType | undefined>(undefined);

export function RoleProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();

  const [activeRole, setActiveRole] = useState<UserRole>("competitor");
  const [organizerProfile, setOrganizerProfile] =
    useState<OrganizerProfile | null>(DEFAULT_ORGANIZER);
  const [sponsorProfile, setSponsorProfile] = useState<SponsorProfile | null>(
    DEFAULT_SPONSOR
  );

  // Compute the display label for the currently active profile
  const activeProfileName = useMemo((): string => {
    if (activeRole === "competitor") {
      return user?.displayName || "Competitor";
    }
    if (activeRole === "organizer") {
      return organizerProfile?.name || "Organizer (Unset)";
    }
    if (activeRole === "sponsor") {
      return sponsorProfile?.companyName || "Sponsor (Unset)";
    }
    return "";
  }, [
    activeRole,
    user?.displayName,
    organizerProfile?.name,
    sponsorProfile?.companyName,
  ]);

  // Switch active contextual role
  const setRole = useCallback((role: UserRole) => {
    setActiveRole(role);
  }, []);

  // Create or update the single organizer profile for this user
  const createOrUpdateOrganizerProfile = useCallback(
    (name: string, bio?: string) => {
      setOrganizerProfile((prev) => ({
        id: prev?.id || `org-${Date.now()}`,
        name,
        bio,
      }));
      setActiveRole("organizer");
    },
    []
  );

  // Create or update the single sponsor profile for this user
  const createOrUpdateSponsorProfile = useCallback(
    (companyName: string, websiteUrl?: string) => {
      setSponsorProfile((prev) => ({
        id: prev?.id || `sp-${Date.now()}`,
        companyName,
        websiteUrl,
      }));
      setActiveRole("sponsor");
    },
    []
  );

  const value = useMemo<RoleContextType>(
    () => ({
      activeRole,
      activeProfileName,
      organizerProfile,
      sponsorProfile,
      setRole,
      createOrUpdateOrganizerProfile,
      createOrUpdateSponsorProfile,
    }),
    [
      activeRole,
      activeProfileName,
      organizerProfile,
      sponsorProfile,
      setRole,
      createOrUpdateOrganizerProfile,
      createOrUpdateSponsorProfile,
    ]
  );

  return <RoleContext.Provider value={value}>{children}</RoleContext.Provider>;
}

export function useRole() {
  const context = useContext(RoleContext);
  if (!context) {
    throw new Error("useRole must be used within a RoleProvider");
  }
  return context;
}
