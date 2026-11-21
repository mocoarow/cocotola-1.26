import { fetchWithIdToken, getAuthUrl } from "./fetch.server";

type ExchangeResult = {
  accessToken: string;
  refreshToken: string;
};

export async function exchangeSupabaseToken(
  supabaseJwt: string,
  organizationName: string,
): Promise<ExchangeResult> {
  console.info(
    `[auth] exchangeSupabaseToken called: organizationName=${organizationName}, jwtLength=${supabaseJwt.length}`,
  );

  const authUrl = getAuthUrl();

  const apiKey = process.env.AUTH_API_KEY;
  if (!apiKey) {
    throw new Error("AUTH_API_KEY environment variable is required");
  }

  const url = `${authUrl}/api/v1/internal/auth/supabase/exchange`;
  console.info(`[auth] calling token exchange: url=${url}`);

  const response = await fetchWithIdToken(authUrl, url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Service-Api-Key": apiKey,
    },
    body: JSON.stringify({ supabaseJwt, organizationName }),
  });

  if (!response.ok) {
    const body = await response.text();
    console.error(`[auth] token exchange failed: status=${response.status}, body=${body}`);
    throw new Error(`Token exchange failed (${response.status}): ${body}`);
  }

  const data = (await response.json()) as ExchangeResult;
  console.info("[auth] token exchange succeeded");
  return data;
}

const REFRESH_TIMEOUT_MS = 3_000;

export type RefreshResult = { kind: "refreshed"; accessToken: string } | { kind: "unauthorized" };

function readAccessToken(data: unknown): string | undefined {
  if (typeof data !== "object" || data === null || !("accessToken" in data)) return undefined;
  return typeof data.accessToken === "string" && data.accessToken.length > 0
    ? data.accessToken
    : undefined;
}

export async function refreshAccessToken(refreshToken: string): Promise<RefreshResult> {
  const authUrl = getAuthUrl();

  const response = await fetchWithIdToken(authUrl, `${authUrl}/api/v1/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refreshToken }),
    signal: AbortSignal.timeout(REFRESH_TIMEOUT_MS),
  });

  if (response.status === 401) {
    console.info("[auth] token refresh rejected: status=401");
    return { kind: "unauthorized" };
  }
  if (!response.ok) {
    throw new Error(`Token refresh failed (${response.status})`);
  }

  const accessToken = readAccessToken(await response.json());
  if (accessToken === undefined) {
    throw new Error("Token refresh returned no access token");
  }
  return { kind: "refreshed", accessToken };
}
