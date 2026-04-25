const configuredApiBase = process.env.NEXT_PUBLIC_API_BASE?.trim() ?? "";

export const apiBaseUrl = configuredApiBase.replace(/\/$/, "");
