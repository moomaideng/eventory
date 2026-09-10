import type { Metadata } from "next";
import { LobbyCodeForm } from "@/features/lobbies/components/lobby-code-form";

export const metadata: Metadata = {
  title: "Join Team Lobby - Eventory",
  description:
    "Enter an invite code from your team captain to view the roster and join the lobby.",
};

export default function LobbiesPage() {
  return <LobbyCodeForm />;
}
