import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/client.server", () => ({
  createBackendClient: vi.fn(),
}));

vi.mock("~/lib/api/study.server", () => ({
  getStudyQuestions: vi.fn(),
  getStudySummary: vi.fn(),
  recordAnswer: vi.fn(),
  deleteStudyHistory: vi.fn(),
  listStudyRecords: vi.fn(),
}));

import { createBackendClient } from "~/lib/api/client.server";
import { listStudyRecords } from "~/lib/api/study.server";
import { backendClientSentinel } from "~/lib/api/test-helpers";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { loader } from "./workbooks.$workbookId.debug-study";

type LoaderArgs = Parameters<typeof loader>[0];

const backendClient = backendClientSentinel();

const RECORD = {
  workbookId: "wb-1",
  questionId: "q-42",
  consecutiveCorrect: 2,
  lastAnsweredAt: "2026-11-20T10:00:00Z",
  nextDueAt: "2026-11-22T10:00:00Z",
  totalCorrect: 5,
  totalIncorrect: 1,
};

function buildLoaderArgs(): LoaderArgs {
  const request = new Request("http://localhost/workbooks/wb-1/debug-study");
  return { request, params: { workbookId: "wb-1" }, context: {} } as unknown as LoaderArgs;
}

describe("DebugStudyRecords loader", () => {
  beforeEach(() => {
    vi.mocked(requireAuth).mockReturnValue({ accessToken: "test-token" });
    vi.mocked(createBackendClient).mockReturnValue(backendClient);
    vi.mocked(listStudyRecords).mockResolvedValue({ records: [RECORD] });
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("should create the backend client with the request the loader received", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].request).toBe(args.request);
  });

  it("should create the backend client with the access token", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].accessToken).toBe("test-token");
  });

  it("should list the study records through the backend client", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(listStudyRecords).mock.calls[0][0]).toBe(backendClient);
  });

  it("should return the workbook id and the study records", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    const result = await loader(args);

    // then
    expect(result).toEqual({ workbookId: "wb-1", records: [RECORD] });
  });
});
