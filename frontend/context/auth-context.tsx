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
import type { Session } from "@supabase/supabase-js";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { MOCK_USER } from "./mock-data";

export interface UserProfile {
  id: string;
  email: string;
  displayName: string;
  handle: string;
  avatarUrl?: string;
  phone?: string;
}

export interface AuthContextType {
  user: UserProfile | null;
  isLoading: boolean;
  authorizationHeader: string | undefined;
  loginWithGoogle: () => Promise<void>;
  loginAsDev: (role?: string) => void;
  logout: () => Promise<void>;
  updateUserProfile: (
    displayName: string,
    handle?: string,
    avatarUrl?: string,
    phone?: string
  ) => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

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
    console.error("[Auth] Google login error:", error);
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [devUser, setDevUser] = useState<UserProfile | null>(null);
  const [session, setSession] = useState<{
    accessToken: string;
    avatarUrl?: string;
    user: {
      id: string;
      email: string;
      displayName: string;
      avatarUrl?: string;
    };
  } | null>(null);
  const [isAuthLoading, setIsAuthLoading] = useState(true);

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

  // Derived user profile from query data or dev user state with zero-flicker OAuth fallback
  const user = useMemo<UserProfile | null>(() => {
    if (devUser) {
      return devUser;
    }
    if (!session?.accessToken || !session.user) {
      return null;
    }
    if (account) {
      return {
        id: account.id,
        email: account.email,
        displayName: account.displayName || session.user.displayName,
        handle: account.handle || session.user.email.split("@")[0],
        avatarUrl: account.avatarUrl || session.user.avatarUrl,
        phone: account.phone || undefined,
      };
    }
    // Instant fallback from Supabase session while backend account query resolves
    return {
      id: session.user.id,
      email: session.user.email,
      displayName: session.user.displayName,
      handle: session.user.email.split("@")[0],
      avatarUrl: session.user.avatarUrl,
    };
  }, [devUser, session, account]);

  const isLoading =
    isAuthLoading ||
    (Boolean(session?.accessToken) && !devUser && isAccountLoading && !account);

  const authorizationHeader = devUser
    ? "Bearer dev-token"
    : session?.accessToken
      ? `Bearer ${session.accessToken}`
      : undefined;

  // Query cache invalidation and refetch on demand / profile update
  const refreshUser = useCallback(async () => {
    await queryClient.invalidateQueries({
      queryKey: ["get", "/api/v1/accounts/me"],
    });
    await refetchAccount();
  }, [queryClient, refetchAccount]);

  useEffect(() => {
    const supabase = createClient();

    const parseSession = (sbSession: Session | null) => {
      if (!sbSession?.access_token || !sbSession.user) {
        return null;
      }
      const sbUser = sbSession.user;
      const userMeta = sbUser.user_metadata || {};
      const displayName =
        userMeta.full_name ||
        userMeta.name ||
        sbUser.email?.split("@")[0] ||
        "Player";
      const avatarUrl = userMeta.avatar_url || userMeta.picture;

      return {
        accessToken: sbSession.access_token,
        avatarUrl: avatarUrl as string | undefined,
        user: {
          id: sbUser.id,
          email: sbUser.email || "",
          displayName: displayName as string,
          avatarUrl: avatarUrl as string | undefined,
        },
      };
    };

    // Initial session check
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(parseSession(session));
      setIsAuthLoading(false);
    });

    // Real-time auth state listener
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((event, session) => {
      if (event === "SIGNED_OUT") {
        setSession(null);
        setDevUser(null);
        queryClient.removeQueries({
          queryKey: ["get", "/api/v1/accounts/me"],
        });
      } else {
        setSession(parseSession(session));
      }
      setIsAuthLoading(false);
    });

    return () => {
      subscription.unsubscribe();
    };
  }, [queryClient]);

  // Instant mock sign-in for zero-friction local development
  const loginAsDev = useCallback((_role?: string) => {
    void _role;
    setDevUser(MOCK_USER);
    if (typeof document !== "undefined") {
      document.cookie =
        "eventory_dev_session=true; path=/; max-age=86400; SameSite=Lax";
    }
  }, []);

  // Update primary user profile details with query cache update & invalidation
  const updateUserProfile = useCallback(
    (
      displayName: string,
      handle?: string,
      avatarUrl?: string,
      phone?: string
    ) => {
      if (devUser) {
        setDevUser((prev) =>
          prev
            ? {
                ...prev,
                displayName,
                handle: handle || prev.handle,
                avatarUrl: avatarUrl || prev.avatarUrl,
                phone: phone || prev.phone,
              }
            : null
        );
        return;
      }

      queryClient.setQueriesData<components["schemas"]["AccountResponse"]>(
        { queryKey: ["get", "/api/v1/accounts/me"] },
        (old) => {
          if (!old) return old;
          return {
            ...old,
            displayName,
            handle: handle || old.handle,
            avatarUrl: avatarUrl || old.avatarUrl,
            phone: phone || old.phone,
          };
        }
      );
      queryClient.invalidateQueries({
        queryKey: ["get", "/api/v1/accounts/me"],
      });
    },
    [devUser, queryClient]
  );

  // Sign out user, reset context state, and redirect to root landing page
  const logout = useCallback(async () => {
    try {
      const supabase = createClient();
      await supabase.auth.signOut();
    } catch {
      // Ignore cleanup error
    }
    if (typeof document !== "undefined") {
      document.cookie = "eventory_dev_session=; path=/; max-age=0";
    }
    setDevUser(null);
    setSession(null);
    queryClient.removeQueries({
      queryKey: ["get", "/api/v1/accounts/me"],
    });
    router.push("/");
    router.refresh();
  }, [queryClient, router]);

  const value = useMemo<AuthContextType>(
    () => ({
      user,
      isLoading,
      authorizationHeader,
      loginWithGoogle,
      loginAsDev,
      logout,
      updateUserProfile,
      refreshUser,
    }),
    [
      user,
      isLoading,
      authorizationHeader,
      loginAsDev,
      logout,
      updateUserProfile,
      refreshUser,
    ]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
