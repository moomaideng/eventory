import type { Metadata } from "next";
import { ModeHub } from "@/features/hub/components/mode-hub";

export const metadata: Metadata = {
  title: "Choose Your Mode - Eventory",
  description:
    "Select what profile mode you want to use to interact with Eventory.",
};

export default function HubPage() {
  return <ModeHub />;
}
