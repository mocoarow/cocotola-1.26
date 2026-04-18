import { getIdToken } from "./id-token.server";

export function getQuestionUrl(): string {
  const url = process.env.QUESTION_BASE_URL;
  if (!url) {
    throw new Error("QUESTION_BASE_URL environment variable is required");
  }
  return url;
}

/**
 * fetch wrapper that attaches a Google Cloud ID Token
 * when running on Cloud Run (APP_ENV is not "local" or "test").
 *
 * @param audience - The Cloud Run service URL used as the ID token audience.
 */
export async function fetchWithIdToken(
  audience: string,
  url: string,
  init?: RequestInit,
): Promise<Response> {
  console.info(
    `[fetch] fetchWithIdToken called: audience=${audience}, method=${init?.method ?? "GET"}, url=${url}`,
  );

  const idToken = await getIdToken(audience);

  if (idToken) {
    const headers = new Headers(init?.headers);
    headers.set("X-Serverless-Authorization", `Bearer ${idToken}`);
    console.info(`[fetch] X-Serverless-Authorization header set (token length=${idToken.length})`);

    const headerKeys = [...headers.keys()].join(", ");
    console.info(`[fetch] request headers: ${headerKeys}`);

    const response = await fetch(url, { ...init, headers });
    console.info(`[fetch] response: status=${response.status}, url=${url}`);
    return response;
  }

  console.info("[fetch] no ID token attached (local/test mode)");
  const response = await fetch(url, init);
  console.info(`[fetch] response: status=${response.status}, url=${url}`);
  return response;
}
