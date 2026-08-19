import type { NextConfig } from "next";

const defaultGoAPIOrigin = "http://127.0.0.1:8081";

function resolveGoAPIOrigin(): string {
  const configuredOrigin = process.env.GO_API_ORIGIN?.trim() || defaultGoAPIOrigin;

  try {
    const origin = new URL(configuredOrigin);
    const hasUnexpectedParts =
      (origin.protocol !== "http:" && origin.protocol !== "https:") ||
      origin.username !== "" ||
      origin.password !== "" ||
      (origin.pathname !== "" && origin.pathname !== "/") ||
      origin.search !== "" ||
      origin.hash !== "";

    if (hasUnexpectedParts) {
      throw new Error("invalid origin");
    }

    return origin.origin;
  } catch {
    // Keep the target server-only and fail startup without echoing a possibly
    // sensitive or malformed environment value into logs.
    throw new Error("GO_API_ORIGIN must be a valid HTTP or HTTPS origin");
  }
}

const goAPIOrigin = resolveGoAPIOrigin();

const nextConfig: NextConfig = {
  // Keep the repository's reviewed root instructions authoritative. Next's
  // generated agent files are development aids, not application artifacts.
  agentRules: false,
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${goAPIOrigin}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
