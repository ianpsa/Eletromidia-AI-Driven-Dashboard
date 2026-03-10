const API_BASE = ((import.meta.env.VITE_BASE_URL as string) || "/api").replace(
  /\/$/,
  "",
);

export function buildApiUrl(
  endpoint: string,
  query: Record<string, string | undefined> = {},
): string {
  const basePath = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  const url = new URL(`${API_BASE}${basePath}`, window.location.origin);

  Object.entries(query).forEach(([key, value]) => {
    if (value) {
      url.searchParams.set(key, value);
    }
  });

  return url.toString();
}
