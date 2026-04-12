import { getIdToken } from "./id-token.server";

/**
 * fetch wrapper that attaches a Google Cloud ID Token
 * when running on Cloud Run (APP_ENV is not "local" or "test").
 */
export async function fetchWithIdToken(
  serviceBaseUrl: string,
  url: string,
  init?: RequestInit,
): Promise<Response> {
  const idToken = await getIdToken(serviceBaseUrl);

  if (idToken) {
    const headers = new Headers(init?.headers);
    headers.set("X-Serverless-Authorization", `Bearer ${idToken}`);
    return fetch(url, { ...init, headers });
  }

  return fetch(url, init);
}
