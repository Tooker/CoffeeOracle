const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE?.trim() ?? "";

const normalizedApiBase = configuredApiBase.replace(/\/$/, "");

const pointsToLocalhost = /^https?:\/\/(localhost|127\.0\.0\.1)(:\d+)?$/i.test(
  normalizedApiBase,
);

export const apiBaseUrl = pointsToLocalhost ? "" : normalizedApiBase;
