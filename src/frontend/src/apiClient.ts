/**
 * Thin wrapper around fetch that:
 *  1. Prepends VITE_BASE_URL to relative paths
 *  2. Injects the Firebase ID token as Authorization: Bearer <token>
 *  3. Re-throws on non-OK responses with a readable message
 *
 * Usage:
 *   const client = useApiClient();
 *   const data   = await client.get("/api/something");
 *   const result = await client.post("/api/other", { body: payload });
 */

import { useCallback } from "react";
import { useAuth } from "./AuthContext";

const API_BASE = ((import.meta.env.VITE_BASE_URL as string) || "").replace(/\/$/, "");

function resolveUrl(endpoint: string): string {
  if (endpoint.startsWith("http")) return endpoint;
  const path = endpoint.startsWith("/") ? endpoint : `/${endpoint}`;
  return `${API_BASE}${path}`;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`[${res.status}] ${text}`);
  }
  const ct = res.headers.get("content-type") ?? "";
  return ct.includes("application/json") ? res.json() : (res.text() as unknown as T);
}

export function useApiClient() {
  const { getToken } = useAuth();

  const request = useCallback(
    async <T>(
      method: string,
      endpoint: string,
      options: { body?: unknown; query?: Record<string, string | undefined> } = {},
    ): Promise<T> => {
      const token = await getToken();
      const url = new URL(resolveUrl(endpoint));

      if (options.query) {
        for (const [k, v] of Object.entries(options.query)) {
          if (v !== undefined) url.searchParams.set(k, v);
        }
      }

      const res = await fetch(url.toString(), {
        method,
        headers: {
          "Authorization": `Bearer ${token}`,
          ...(options.body ? { "Content-Type": "application/json" } : {}),
        },
        ...(options.body ? { body: JSON.stringify(options.body) } : {}),
      });

      return handleResponse<T>(res);
    },
    [getToken],
  );

  return {
    get:  <T>(endpoint: string, query?: Record<string, string | undefined>) =>
            request<T>("GET", endpoint, { query }),
    post: <T>(endpoint: string, body?: unknown) =>
            request<T>("POST", endpoint, { body }),
    put:  <T>(endpoint: string, body?: unknown) =>
            request<T>("PUT", endpoint, { body }),
    del:  <T>(endpoint: string) =>
            request<T>("DELETE", endpoint),
  };
}