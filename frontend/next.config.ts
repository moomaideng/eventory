import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  // These values are intentionally public and embedded in the browser bundle.
  env: {
    SUPABASE_URL: process.env.SUPABASE_URL,
    SUPABASE_PUBLISHABLE_KEY: process.env.SUPABASE_PUBLISHABLE_KEY,
    API_URL: process.env.API_URL,
  },
};

export default nextConfig;
