import type { NextConfig } from "next";
import withBundleAnalyzer from "@next/bundle-analyzer";

const nextConfig: NextConfig = {
  // @ts-ignore - allowedDevOrigins is supported but might be missing from NextConfig type in this version
  allowedDevOrigins: ['juliyenglish.id.vn', '*.juliyenglish.id.vn'],
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://127.0.0.1:8080/api/v1/:path*', // Proxy to Backend
      },
    ];
  },
};

const withAnalyzer = withBundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
});

export default withAnalyzer(nextConfig);
