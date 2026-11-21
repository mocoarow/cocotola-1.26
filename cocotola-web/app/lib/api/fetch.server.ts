import { SpanStatusCode } from "@opentelemetry/api";
import { isServerErrorStatus } from "~/lib/http-status";
import { withSpan } from "~/lib/observability/tracing.server";
import { getIdToken } from "./id-token.server";

export function getQuestionUrl(): string {
  const url = process.env.QUESTION_BASE_URL;
  if (!url) {
    throw new Error("QUESTION_BASE_URL environment variable is required");
  }
  return url;
}

export function getAuthUrl(): string {
  const url = process.env.AUTH_BASE_URL;
  if (!url) {
    throw new Error("AUTH_BASE_URL environment variable is required");
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
      const idToken = await getIdToken(audience, init?.signal ?? undefined);

      let response: Response;
      if (idToken) {
        const headers = new Headers(init?.headers);
        headers.set("X-Serverless-Authorization", `Bearer ${idToken}`);
        response = await fetch(url, { ...init, headers });
      } else {
        response = await fetch(url, init);
      }

      span.setAttribute("http.response.status_code", response.status);
      if (isServerErrorStatus(response.status)) {
        span.setStatus({ code: SpanStatusCode.ERROR });
      }
      return response;
    },
    {
      "http.request.method": method,
      "url.full": url,
      "server.address": audience,
    },
  );
}
