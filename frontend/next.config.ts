import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  allowedDevOrigins: ["127.0.0.1"],
  async redirects() {
    return [
      {
        source: "/tournament/new",
        destination: "/organizer/tournaments/new",
        permanent: false,
      },
      {
        source: "/tournaments/new",
        destination: "/organizer/tournaments/new",
        permanent: false,
      },
    ];
  },
};

export default nextConfig;
