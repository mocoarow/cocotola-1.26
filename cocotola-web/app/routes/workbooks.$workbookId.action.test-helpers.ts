import { afterEach, beforeEach, vi } from "vitest";
import { createBackendClient } from "~/lib/api/client.server";
import { backendClientSentinel } from "~/lib/api/test-helpers";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { action } from "./workbooks.$workbookId";

export type ActionArgs = Parameters<typeof action>[0];

export const ACCESS_TOKEN = "test-token";

export const backendClient = backendClientSentinel();

export function buildActionArgs(formEntries: Record<string, string>): ActionArgs {
  const request = new Request("http://localhost/workbooks/wb-1", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams(formEntries).toString(),
  });
  return { request, params: { workbookId: "wb-1" }, context: {} } as unknown as ActionArgs;
}

export function setupActionMocks() {
  beforeEach(() => {
    vi.mocked(requireAuth).mockReturnValue({ accessToken: ACCESS_TOKEN });
    vi.mocked(createBackendClient).mockReturnValue(backendClient);
  });

  afterEach(() => {
    vi.resetAllMocks();
  });
}
