"use client";

import React from "react";
import { useRouter } from "next/navigation";
import {
  Bell,
  CheckCircle2,
  XCircle,
  Receipt,
  CheckCheck,
  BellOff,
  ExternalLink,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useNotifications } from "@/features/notifications/hooks/use-notifications";
import type { NotificationItem } from "@/features/notifications/types";
import { cn } from "@/lib/utils";

function formatRelativeTime(dateInput: string): string {
  const date = new Date(dateInput);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return "just now";
  const diffInMinutes = Math.floor(diffInSeconds / 60);
  if (diffInMinutes < 60) return `${diffInMinutes}m ago`;
  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) return `${diffInHours}h ago`;
  const diffInDays = Math.floor(diffInHours / 24);
  if (diffInDays < 7) return `${diffInDays}d ago`;
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function getNotificationIcon(type: string) {
  switch (type) {
    case "REGISTRATION_APPROVED":
      return (
        <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
          <CheckCircle2 className="size-4" />
        </div>
      );
    case "REGISTRATION_REJECTED":
      return (
        <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-destructive/10 text-destructive">
          <XCircle className="size-4" />
        </div>
      );
    case "PAYMENT_CONFIRMED":
      return (
        <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-blue-500/10 text-blue-600 dark:text-blue-400">
          <Receipt className="size-4" />
        </div>
      );
    default:
      return (
        <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
          <Bell className="size-4" />
        </div>
      );
  }
}

export function NotificationBell() {
  const router = useRouter();
  const {
    notifications,
    unreadCount,
    isLoading,
    refetch,
    markAsRead,
    markAllAsRead,
  } = useNotifications();

  const handleNotificationClick = async (item: NotificationItem) => {
    if (!item.isRead) {
      await markAsRead(item.id);
    }
    if (item.link) {
      router.push(item.link);
    }
  };

  return (
    <DropdownMenu
      onOpenChange={(open) => {
        if (open) void refetch();
      }}
    >
      <DropdownMenuTrigger
        render={
          <Button
            variant="outline"
            size="icon"
            className="relative h-9 w-9 rounded-lg"
            aria-label={`Notifications (${unreadCount} unread)`}
            onClick={() => void refetch()}
          />
        }
      >
        <Bell className="size-4" />
        {unreadCount > 0 && (
          <span className="absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground shadow-xs animate-in zoom-in-75">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-80 sm:w-96 rounded-xl border border-border bg-popover p-0 shadow-xl"
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <div className="flex items-center gap-2">
            <h3 className="text-sm font-semibold text-foreground">
              Notifications
            </h3>
            {unreadCount > 0 && (
              <Badge variant="default" className="text-[10px] h-4 px-1.5">
                {unreadCount} new
              </Badge>
            )}
          </div>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              className="h-7 px-2 text-xs text-muted-foreground hover:text-foreground"
              onClick={(e) => {
                e.stopPropagation();
                void markAllAsRead();
              }}
            >
              <CheckCheck className="size-3.5 mr-1" />
              Mark all read
            </Button>
          )}
        </div>

        {/* Content list */}
        <div className="max-h-[380px] overflow-y-auto divide-y divide-border/50">
          {isLoading ? (
            <div className="p-4 space-y-3">
              <Skeleton className="h-12 w-full rounded-lg" />
              <Skeleton className="h-12 w-full rounded-lg" />
              <Skeleton className="h-12 w-full rounded-lg" />
            </div>
          ) : notifications.length === 0 ? (
            <div className="flex flex-col items-center justify-center px-4 py-10 text-center">
              <div className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground mb-3">
                <BellOff className="size-5 opacity-60" />
              </div>
              <p className="text-sm font-medium text-foreground">
                No notifications
              </p>
              <p className="text-xs text-muted-foreground mt-1 max-w-[200px]">
                You have no notifications right now.
              </p>
            </div>
          ) : (
            notifications.map((item) => (
              <div
                key={item.id}
                role="button"
                tabIndex={0}
                onClick={() => void handleNotificationClick(item)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    void handleNotificationClick(item);
                  }
                }}
                className={cn(
                  "group flex items-start gap-3 p-3.5 transition-colors cursor-pointer text-left hover:bg-muted/50 focus-visible:bg-muted/50 focus-visible:outline-none",
                  !item.isRead && "bg-primary/5 dark:bg-primary/10"
                )}
              >
                {getNotificationIcon(item.type)}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-1 mb-0.5">
                    <p
                      className={cn(
                        "text-xs font-semibold truncate",
                        !item.isRead ? "text-foreground font-bold" : "text-muted-foreground"
                      )}
                    >
                      {item.title}
                    </p>
                    <span className="text-[10px] text-muted-foreground shrink-0">
                      {formatRelativeTime(item.createdAt)}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground line-clamp-2 leading-relaxed">
                    {item.message}
                  </p>
                  {item.link && (
                    <span className="inline-flex items-center gap-1 text-[11px] font-medium text-primary mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      View details
                      <ExternalLink className="size-2.5" />
                    </span>
                  )}
                </div>
                {!item.isRead && (
                  <span
                    className="size-2 rounded-full bg-primary shrink-0 self-center"
                    aria-label="Unread notification"
                  />
                )}
              </div>
            ))
          )}
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
