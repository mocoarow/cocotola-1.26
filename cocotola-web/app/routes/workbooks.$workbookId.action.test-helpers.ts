import { afterEach, beforeEach, vi } from "vitest";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { action } from "./workbooks.$workbookId";

export type ActionArgs = Parameters<typeof action>[0];

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
    vi.mocked(requireAuth).mockReturnValue({ accessToken: "test-token" });
  });

  afterEach(() => {
    vi.resetAllMocks();
  });
}
