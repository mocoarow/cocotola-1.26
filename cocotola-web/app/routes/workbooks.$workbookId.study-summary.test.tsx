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
import { getStudySummary } from "~/lib/api/study.server";
import { backendClientSentinel } from "~/lib/api/test-helpers";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { loader } from "./workbooks.$workbookId.study-summary";

type LoaderArgs = Parameters<typeof loader>[0];

const backendClient = backendClientSentinel();

const SUMMARY = {
  newCount: 5,
  reviewCount: 12,
  totalDue: 17,
  reviewRatioNumerator: 7,
  reviewRatioDenominator: 11,
};

function buildLoaderArgs(url = "http://localhost/workbooks/wb-1/study-summary"): LoaderArgs {
  const request = new Request(url);
  return { request, params: { workbookId: "wb-1" }, context: {} } as unknown as LoaderArgs;
}

describe("StudySummary loader", () => {
  beforeEach(() => {
    vi.mocked(requireAuth).mockReturnValue({ accessToken: "test-token" });
    vi.mocked(createBackendClient).mockReturnValue(backendClient);
    vi.mocked(getStudySummary).mockResolvedValue(SUMMARY);
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

  it("should fetch the summary through the backend client", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(getStudySummary).mock.calls[0][0]).toBe(backendClient);
  });

  it("should request the practice summary when practice=true is in the URL", async () => {
    // given
    const args = buildLoaderArgs("http://localhost/workbooks/wb-1/study-summary?practice=true");

    // when
    await loader(args);

    // then
    expect(vi.mocked(getStudySummary).mock.calls[0][2]).toBe(true);
  });

  it("should return the summary", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    const result = await loader(args);

    // then
    expect(result).toEqual(SUMMARY);
  });
});
