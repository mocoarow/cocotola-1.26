import { fetchWithIdToken } from "./fetch.server";

type ExchangeResult = {
  accessToken: string;
  refreshToken: string;
};

export async function exchangeSupabaseToken(
  supabaseJwt: string,
  organizationName: string,
): Promise<ExchangeResult> {
  const authUrl = process.env.AUTH_BASE_URL;
  if (!authUrl) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }

  const apiKey = process.env.AUTH_API_KEY;
  if (!apiKey) {
    throw new Error("AUTH_API_KEY environment variable is required");
  }

  const response = await fetchWithIdToken(
    "cocotola-auth",
    `${authUrl}/api/v1/internal/auth/supabase/exchange`,
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
    console.error(
      `[api] POST ${authUrl}/api/v1/internal/auth/supabase/exchange -> ${response.status}: ${body}`,
    );
    throw new Error(`Token exchange failed (${response.status}): ${body}`);
  }

  const data = (await response.json()) as ExchangeResult;
  return data;
}
