"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { createClient } from "@/lib/client";
import { apiClient } from "@/lib/api/client";

export type UserRole = "competitor" | "organizer" | "sponsor";

export interface UserProfile {
  id: string;
  email: string;
  displayName: string;
  avatarUrl?: string;
}

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

interface RoleContextType {
  user: UserProfile | null;
  activeRole: UserRole;
  activeProfileName: string;
  organizerProfile: OrganizerProfile | null;
  sponsorProfile: SponsorProfile | null;
  isLoading: boolean;
  setRole: (role: UserRole) => void;
  loginWithGoogle: () => Promise<void>;
  loginAsDev: (role?: UserRole) => void;
  logout: () => Promise<void>;
  updateUserProfile: (displayName: string, avatarUrl?: string) => void;
  createOrUpdateOrganizerProfile: (name: string, bio?: string) => void;
  createOrUpdateSponsorProfile: (
    companyName: string,
    websiteUrl?: string
  ) => void;
  refreshUser: () => Promise<void>;
}

const RoleContext = createContext<RoleContextType | undefined>(undefined);

// Mock data for local development mode
const MOCK_USER: UserProfile = {
  id: "dev-user-001",
  email: "dev@eventory.gg",
  displayName: "MooMai (Dev)",
  avatarUrl: "https://api.dicebear.com/7.x/avataaars/svg?seed=MooMai",
};

const DEFAULT_ORGANIZER: OrganizerProfile = {
  id: "org-1",
  name: "Chula Esports Club",
  bio: "Official university esports club hosting regional tournaments.",
};

const DEFAULT_SPONSOR: SponsorProfile = {
  id: "sp-1",
  companyName: "Red Bull Gaming",
  websiteUrl: "https://redbull.com",
};

export function RoleProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(null);
  const [activeRole, setActiveRole] = useState<UserRole>("competitor");
  const [organizerProfile, setOrganizerProfile] =
    useState<OrganizerProfile | null>(DEFAULT_ORGANIZER);
  const [sponsorProfile, setSponsorProfile] = useState<SponsorProfile | null>(
    DEFAULT_SPONSOR
  );
  const [isLoading, setIsLoading] = useState(true);

  // Sync user state with Supabase session and Go Backend via openapi-fetch
  const refreshUser = async () => {
    try {
      const supabase = createClient();
      const {
        data: { session },
      } = await supabase.auth.getSession();

      if (session?.access_token) {
        // Query Go Backend Account API directly with openapi-fetch
        const { data: account } = await apiClient.GET(
          "/api/v1/accounts/me",
          {
            headers: {
              Authorization: `Bearer ${session.access_token}`,
            },
          }
        );

        if (account) {
          setUser({
            id: account.id,
            email: account.email,
            displayName: account.username,
            avatarUrl: session.user?.user_metadata?.avatar_url,
          });
        } else {
          // Account does not exist in DB yet (needs onboarding)
          setUser(null);
        }
      } else {
        setUser(null);
      }
    } catch (err) {
      if (process.env.NODE_ENV === "development") {
        console.log("Auth session check notice:", err);
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const supabase = createClient();

    // 1. Initial auth sync
    refreshUser();

    // 2. Real-time auth state listener (Sign in, Sign out, Token Refresh)
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange(async (event, session) => {
      if (session?.access_token) {
        try {
          const { data: account } = await apiClient.GET(
            "/api/v1/accounts/me",
            {
              headers: {
                Authorization: `Bearer ${session.access_token}`,
              },
            }
          );

          if (account) {
            setUser({
              id: account.id,
              email: account.email,
              displayName: account.username,
              avatarUrl: session.user?.user_metadata?.avatar_url,
            });
          } else {
            setUser(null);
          }
        } catch {
          setUser(null);
        }
      } else if (event === "SIGNED_OUT") {
        setUser(null);
        setActiveRole("competitor");
      }
      setIsLoading(false);
    });

    return () => {
      subscription.unsubscribe();
    };
  }, []);

  // Compute the display label for the currently active profile
  const getActiveProfileName = (): string => {
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
  };

  // Switch active contextual role
  const setRole = (role: UserRole) => {
    setActiveRole(role);
  };

  // Trigger Supabase Google OAuth sign-in
  const loginWithGoogle = async () => {
    try {
      const supabase = createClient();
      await supabase.auth.signInWithOAuth({
        provider: "google",
        options: {
          redirectTo: `${window.location.origin}/api/auth/callback`,
        },
      });
    } catch (error) {
      console.error("Google login error:", error);
    }
  };

  // Instant mock sign-in for zero-friction local development
  const loginAsDev = (role: UserRole = "competitor") => {
    setUser(MOCK_USER);
    setActiveRole(role);
    setOrganizerProfile(DEFAULT_ORGANIZER);
    setSponsorProfile(DEFAULT_SPONSOR);
  };

  // Update primary user profile details
  const updateUserProfile = (displayName: string, avatarUrl?: string) => {
    if (!user) {
      setUser({
        id: `user-${Date.now()}`,
        email: "user@eventory.gg",
        displayName,
        avatarUrl,
      });
    } else {
      setUser((prev) =>
        prev
          ? { ...prev, displayName, avatarUrl: avatarUrl || prev.avatarUrl }
          : null
      );
    }
  };

  // Sign out user and reset context state
  const logout = async () => {
    try {
      const supabase = createClient();
      await supabase.auth.signOut();
    } catch {
      // Ignore cleanup error
    }
    setUser(null);
    setActiveRole("competitor");
  };

  // Create or update the single organizer profile for this user
  const createOrUpdateOrganizerProfile = (name: string, bio?: string) => {
    const updated: OrganizerProfile = {
      id: organizerProfile?.id || `org-${Date.now()}`,
      name,
      bio,
    };
    setOrganizerProfile(updated);
    setActiveRole("organizer");
  };

  // Create or update the single sponsor profile for this user
  const createOrUpdateSponsorProfile = (
    companyName: string,
    websiteUrl?: string
  ) => {
    const updated: SponsorProfile = {
      id: sponsorProfile?.id || `sp-${Date.now()}`,
      companyName,
      websiteUrl,
    };
    setSponsorProfile(updated);
    setActiveRole("sponsor");
  };

  return (
    <RoleContext.Provider
      value={{
        user,
        activeRole,
        activeProfileName: getActiveProfileName(),
        organizerProfile,
        sponsorProfile,
        isLoading,
        setRole,
        loginWithGoogle,
        loginAsDev,
        logout,
        updateUserProfile,
        createOrUpdateOrganizerProfile,
        createOrUpdateSponsorProfile,
        refreshUser,
      }}
    >
      {children}
    </RoleContext.Provider>
  );
}

export function useRole() {
  const context = useContext(RoleContext);
  if (!context) {
    throw new Error("useRole must be used within a RoleProvider");
  }
  return context;
}
