import { type BackendClient, createBackendClient } from "./client.server";

export function newClient(): BackendClient {
  return createBackendClient({
    request: new Request("http://localhost/some-loader"),
    accessToken: "test-token",
  });
}

export function backendClientSentinel(): BackendClient {
  return {} as BackendClient;
}

export function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

export function unauthorizedResponse(): Response {
  return new Response(null, { status: 401 });
}

export async function catchError(promise: Promise<unknown>): Promise<unknown> {
  return promise.then(
    () => undefined,
    (e: unknown) => e,
  );
}
