import { withSpan } from "~/lib/observability/tracing.server";
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
  const method = init?.method ?? "GET";

  return withSpan(
    `backend ${method}`,
    async (span) => {
      console.info(
        `[fetch] fetchWithIdToken called: audience=${audience}, method=${method}, url=${url}`,
      );

      const idToken = await getIdToken(audience);

      let response: Response;
      if (idToken) {
        const headers = new Headers(init?.headers);
        headers.set("X-Serverless-Authorization", `Bearer ${idToken}`);
        console.info(
          `[fetch] X-Serverless-Authorization header set (token length=${idToken.length})`,
        );

        const headerKeys = [...headers.keys()].join(", ");
        console.info(`[fetch] request headers: ${headerKeys}`);

        response = await fetch(url, { ...init, headers });
      } else {
        console.info("[fetch] no ID token attached (local/test mode)");
        response = await fetch(url, init);
      }

      console.info(`[fetch] response: status=${response.status}, url=${url}`);
      span.setAttribute("http.response.status_code", response.status);
      return response;
    },
    {
      "http.request.method": method,
      "url.full": url,
      "server.address": audience,
    },
  );
}
