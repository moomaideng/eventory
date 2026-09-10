/**
 * Sanitizes a redirect path to ensure it is a safe relative URL within the application.
 * Prevents Open Redirect attacks by rejecting absolute URLs, protocol-relative URLs (//), and external schemes.
 */
export function getSafeRedirectPath(
  target: string | null | undefined,
  fallback = "/hub"
): string {
  if (!target) return fallback;

  const trimmed = target.trim();

  // Must start with a single "/" and NOT followed by another "/" or "\"
  if (
    trimmed.startsWith("/") &&
    !trimmed.startsWith("//") &&
    !trimmed.startsWith("/\\")
  ) {
    return trimmed;
  }

  return fallback;
}

/**
 * Derives a user-friendly destination label to display on the login screen.
 */
export function getRedirectDestinationLabel(path: string): string | null {
  const safePath = getSafeRedirectPath(path);
  if (safePath === "/hub") return null;

  // Match lobby invite code e.g. /lobbies/ABC123
  const lobbyMatch = safePath.match(/^\/lobbies\/([A-Za-z0-9_-]+)/);
  if (lobbyMatch) {
    return `Team Lobby (${lobbyMatch[1].toUpperCase()})`;
  }
  if (safePath === "/lobbies" || safePath.startsWith("/lobbies?")) {
    return "Join Team Lobby";
  }

  // Match tournament team create e.g. /tournaments/123/team
  if (safePath.includes("/team")) {
    return "Team Registration";
  }

  // Workspaces & Hubs
  if (safePath.startsWith("/organizer")) return "Organizer Workspace";
  if (safePath.startsWith("/sponsor")) return "Sponsor Workspace";
  if (safePath.startsWith("/tournaments")) return "Tournaments";
  if (safePath.startsWith("/settings")) return "Account Settings";

  return "your requested page";
}
