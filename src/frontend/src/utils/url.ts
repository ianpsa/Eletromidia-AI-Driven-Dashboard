const API_BASE = ((import.meta.env.VITE_BASE_URL as string) || "").replace(
  /\/$/,
  "",
);

export function buildApiUrl(
  endpoint: string,
  query: Record<string, string | undefined> = {},
): string {
  const basePath = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;

  const baseUrl = API_BASE || window.location.origin;
  const url = new URL(`${baseUrl}${basePath}`);

  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined) {
      url.searchParams.set(key, value);
    }
  });

  console.log("[buildApiUrl] API_BASE:", API_BASE);
  console.log("[buildApiUrl] URL final:", url.toString());

  return url.toString();
}