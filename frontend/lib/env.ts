// configuredApiBase is the raw public API base URL from environment (if provided).
const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE?.trim() ?? "";

// normalizedApiBase removes trailing slash so URL concatenation stays predictable.
const normalizedApiBase = configuredApiBase.replace(/\/$/, "");

// pointsToLocalhost detects local origins; in that case Next rewrites can proxy relatively.
const pointsToLocalhost = /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/i.test(
  normalizedApiBase,
);

// apiBaseUrl is exported as "" for localhost to use relative /api routes via next.config rewrites.
export const apiBaseUrl = pointsToLocalhost ? "" : normalizedApiBase;
