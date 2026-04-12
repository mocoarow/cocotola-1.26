import { fetchWithIdToken } from "./fetch.server";

type ExchangeResult = {
  accessToken: string;
  refreshToken: string;
};

export async function exchangeSupabaseToken(
  supabaseJwt: string,
  organizationName: string,
): Promise<ExchangeResult> {
  console.info(`[auth] exchangeSupabaseToken called: organizationName=${organizationName}, jwtLength=${supabaseJwt.length}`);

  const authUrl = process.env.AUTH_BASE_URL;
  if (!authUrl) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }

  const apiKey = process.env.AUTH_API_KEY;
  if (!apiKey) {
    throw new Error("AUTH_API_KEY environment variable is required");
  }

  const url = `${authUrl}/api/v1/internal/auth/supabase/exchange`;
  console.info(`[auth] calling token exchange: url=${url}`);

  const response = await fetchWithIdToken(
    authUrl,
    url,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Service-Api-Key": apiKey,
      },
      body: JSON.stringify({ supabaseJwt, organizationName }),
    },
  );

  if (!response.ok) {
    const body = await response.text();
    console.error(`[auth] token exchange failed: status=${response.status}, body=${body}`);
    throw new Error(`Token exchange failed (${response.status}): ${body}`);
  }

  const data = (await response.json()) as ExchangeResult;
  console.info("[auth] token exchange succeeded");
  return data;
}
