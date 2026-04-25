import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // rewrites proxies browser requests from Next.js to the Go backend during local development.
  // Users still call /api/* in the frontend; Next forwards to :8080 behind the scenes.
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://127.0.0.1:8080/api/:path*",
      },
    ];
  },
};

export default nextConfig;
