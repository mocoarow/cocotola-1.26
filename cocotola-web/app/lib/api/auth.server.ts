import { fetchWithIdToken } from "./fetch.server";

type ExchangeResult = {
  accessToken: string;
  refreshToken: string;
};

export async function exchangeSupabaseToken(
  supabaseJwt: string,
  organizationName: string,
): Promise<ExchangeResult> {
  const authUrl = process.env.COCOTOLA_AUTH_URL;
  if (!authUrl) {
    throw new Error("COCOTOLA_AUTH_URL environment variable is required");
  }

  const apiKey = process.env.INTERNAL_API_KEY;
  if (!apiKey) {
    throw new Error("INTERNAL_API_KEY environment variable is required");
  }

  const response = await fetchWithIdToken(
    authUrl,
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
