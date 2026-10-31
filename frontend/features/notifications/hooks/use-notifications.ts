"use client";

import { useQueryClient } from "@tanstack/react-query";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import type { NotificationItem } from "@/features/notifications/types";

export function useNotifications() {
  const { user, authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const { data, isLoading, isError, refetch } = $api.useQuery(
    "get",
    "/api/v1/notifications",
    {
      headers: {
        Authorization: authorizationHeader ?? "",
      },
    },
    {
      enabled: Boolean(user && authorizationHeader),
      refetchInterval: 5000,
    }
  );

  const notifications: NotificationItem[] = data?.notifications ?? [];
  const unreadCount: number = data?.unreadCount ?? 0;

  const markAsRead = async (id: string) => {
    if (!authorizationHeader) return;
    try {
      await apiClient.PATCH("/api/v1/notifications/{id}/read", {
        params: {
          path: { id },
        },
        headers: {
          Authorization: authorizationHeader,
        },
      });
      queryClient.invalidateQueries({
        predicate: (query) => query.queryKey[1] === "/api/v1/notifications",
      });
    } catch (err) {
      console.error("Failed to mark notification as read", err);
    }
  };

  const markAllAsRead = async () => {
    if (!authorizationHeader) return;
    try {
      await apiClient.PATCH("/api/v1/notifications/read-all", {
        headers: {
          Authorization: authorizationHeader,
        },
      });
      queryClient.invalidateQueries({
        predicate: (query) => query.queryKey[1] === "/api/v1/notifications",
      });
    } catch (err) {
      console.error("Failed to mark all notifications as read", err);
    }
  };

  return {
    notifications,
    unreadCount,
    isLoading,
    isError,
    refetch,
    markAsRead,
    markAllAsRead,
  };
}
