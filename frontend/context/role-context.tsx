"use client";

import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useMemo,
} from "react";
import { createClient } from "@/lib/client";
import { $api } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
import { useQueryClient } from "@tanstack/react-query";

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

// Pure function at module scope: triggers Supabase Google OAuth sign-in
async function loginWithGoogle() {
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
}

export function RoleProvider({ children }: { children: React.ReactNode }) {
  const queryClient = useQueryClient();

  const [devUser, setDevUser] = useState<UserProfile | null>(null);
  const [session, setSession] = useState<{
    accessToken?: string;
    avatarUrl?: string;
  } | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(true);

  const [activeRole, setActiveRole] = useState<UserRole>("competitor");
  const [organizerProfile, setOrganizerProfile] =
    useState<OrganizerProfile | null>(DEFAULT_ORGANIZER);
  const [sponsorProfile, setSponsorProfile] = useState<SponsorProfile | null>(
    DEFAULT_SPONSOR
  );

  // TanStack Query integration via openapi-react-query client ($api)
  const {
    data: account,
    isLoading: isAccountLoading,
    refetch: refetchAccount,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/me",
    {
      headers: {
        Authorization: session?.accessToken
          ? `Bearer ${session.accessToken}`
          : "",
      },
    },
    {
      enabled: Boolean(session?.accessToken) && !devUser,
      retry: false,
      staleTime: 60 * 1000,
    }
  );

  // Derived user profile from query data or dev user state
  const user = useMemo<UserProfile | null>(() => {
    if (devUser) {
      return devUser;
    }
    if (!session?.accessToken || !account) {
      return null;
    }
    return {
      id: account.id,
      email: account.email,
      displayName: account.username,
      avatarUrl: session.avatarUrl,
    };
  }, [devUser, session, account]);

  const isLoading =
    isAuthLoading ||
    (Boolean(session?.accessToken) && !devUser && isAccountLoading);

  // Query cache invalidation and refetch on demand / onboarding completion
  const refreshUser = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: ["get", "/api/v1/accounts/me"],
    });
    await refetchAccount();
  }, [queryClient, refetchAccount]);

  useEffect(() => {
    const supabase = createClient();

    // Initial session check
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session?.access_token) {
        setSession({
          accessToken: session.access_token,
          avatarUrl: session.user?.user_metadata?.avatar_url,
        });
      } else {
        setSession(null);
      }
      setIsAuthLoading(false);
    });

    // Real-time auth state listener
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((event, session) => {
      if (session?.access_token) {
        setSession({
          accessToken: session.access_token,
          avatarUrl: session.user?.user_metadata?.avatar_url,
        });
      } else if (event === "SIGNED_OUT") {
        setSession(null);
        setDevUser(null);
        setActiveRole("competitor");
        queryClient.removeQueries({
          queryKey: ["get", "/api/v1/accounts/me"],
        });
      } else {
        setSession(null);
      }
      setIsAuthLoading(false);
    });

    return () => {
      subscription.unsubscribe();
    };
  }, [queryClient]);

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

  // Instant mock sign-in for zero-friction local development
  const loginAsDev = useCallback((role: UserRole = "competitor") => {
    setDevUser(MOCK_USER);
    setActiveRole(role);
    setOrganizerProfile(DEFAULT_ORGANIZER);
    setSponsorProfile(DEFAULT_SPONSOR);
  }, []);

  // Update primary user profile details with query cache update & invalidation
  const updateUserProfile = useCallback(
    (displayName: string, avatarUrl?: string) => {
      if (devUser) {
        setDevUser((prev) =>
          prev
            ? { ...prev, displayName, avatarUrl: avatarUrl || prev.avatarUrl }
            : null
        );
        return;
      }

      queryClient.setQueriesData<components["schemas"]["AccountResponse"]>(
        { queryKey: ["get", "/api/v1/accounts/me"] },
        (old) => {
          if (!old) return old;
          return { ...old, username: displayName };
        }
      );
      queryClient.invalidateQueries({
        queryKey: ["get", "/api/v1/accounts/me"],
      });
    },
    [devUser, queryClient]
  );

  // Sign out user and reset context & query state
  const logout = useCallback(async () => {
    try {
      const supabase = createClient();
      await supabase.auth.signOut();
    } catch {
      // Ignore cleanup error
    }
    setDevUser(null);
    setSession(null);
    queryClient.removeQueries({
      queryKey: ["get", "/api/v1/accounts/me"],
    });
    setActiveRole("competitor");
  }, [queryClient]);

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
      user,
      activeRole,
      activeProfileName,
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
    }),
    [
      user,
      activeRole,
      activeProfileName,
      organizerProfile,
      sponsorProfile,
      isLoading,
      setRole,
      loginAsDev,
      logout,
      updateUserProfile,
      createOrUpdateOrganizerProfile,
      createOrUpdateSponsorProfile,
      refreshUser,
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
