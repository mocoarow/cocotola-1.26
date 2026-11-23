import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  listPublicWorkbooks,
  listSharedWorkbooks,
  type PublicWorkbook,
  type SharedReference,
  type SharedWorkbook,
  shareWorkbook,
  unshareWorkbook,
} from "./sharing.server";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";

function buildPublicWorkbook(overrides: Partial<PublicWorkbook> = {}): PublicWorkbook {
  return {
    workbookId: "wb-7",
    ownerId: "user-3",
    title: "French Verbs",
    description: "Irregular verbs in the present tense",
    language: "fr",
    createdAt: "2026-03-04T05:06:07Z",
    ...overrides,
  };
}

function buildSharedWorkbook(overrides: Partial<SharedWorkbook> = {}): SharedWorkbook {
  return {
    referenceId: "ref-5",
    workbookId: "wb-7",
    title: "TOEIC 600",
    description: "Vocabulary for TOEIC 600",
    language: "en",
    addedAt: "2026-04-10T11:12:13Z",
    ...overrides,
  };
}

function buildReference(overrides: Partial<SharedReference> = {}): SharedReference {
  return {
    referenceId: "ref-new",
    workbookId: "wb-7",
    addedAt: "2026-04-29T08:09:10Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.stubEnv("QUESTION_BASE_URL", QUESTION_BASE_URL);
  vi.spyOn(console, "info").mockImplementation(() => {});
  vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
  fetchMock.mockReset();
});

describe("listPublicWorkbooks", () => {
  it("should throw when QUESTION_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("QUESTION_BASE_URL", "");

    // when / then
    await expect(listPublicWorkbooks(newClient())).rejects.toThrow(
      "QUESTION_BASE_URL environment variable is required",
    );
  });

  it("should return public workbooks on success", async () => {
    // given
    const workbooks = [buildPublicWorkbook()];
    fetchMock.mockResolvedValue(jsonResponse({ workbooks }));

    // when
    const result = await listPublicWorkbooks(newClient());

    // then
    expect(result).toEqual(workbooks);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [] }));

    // when
    await listPublicWorkbooks(newClient());

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/public`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when workbooks is undefined", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const result = await listPublicWorkbooks(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should return empty array when workbooks is null", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: null }));

    // when
    const result = await listPublicWorkbooks(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(listPublicWorkbooks(newClient())).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(listPublicWorkbooks(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a workbook in the response lacks ownerId", async () => {
    // given
    const { ownerId: _, ...withoutOwner } = buildPublicWorkbook();
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [withoutOwner] }));

    // when / then
    await expect(listPublicWorkbooks(newClient())).rejects.toMatchObject({ status: 502 });
  });
});

describe("listSharedWorkbooks", () => {
  it("should return shared references on success", async () => {
    // given
    const references = [buildSharedWorkbook()];
    fetchMock.mockResolvedValue(jsonResponse({ references }));

    // when
    const result = await listSharedWorkbooks(newClient());

    // then
    expect(result).toEqual(references);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ references: [] }));

    // when
    await listSharedWorkbooks(newClient());

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/shared`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when references is undefined", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const result = await listSharedWorkbooks(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should return empty array when references is null", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ references: null }));

    // when
    const result = await listSharedWorkbooks(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 403 }));

    // when / then
    await expect(listSharedWorkbooks(newClient())).rejects.toMatchObject({ status: 403 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(listSharedWorkbooks(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a reference in the response lacks title", async () => {
    // given
    const { title: _, ...withoutTitle } = buildSharedWorkbook();
    fetchMock.mockResolvedValue(jsonResponse({ references: [withoutTitle] }));

    // when / then
    await expect(listSharedWorkbooks(newClient())).rejects.toMatchObject({ status: 502 });
  });
});

describe("shareWorkbook", () => {
  it("should return the new reference on success", async () => {
    // given
    const reference = buildReference();
    fetchMock.mockResolvedValue(jsonResponse(reference, 201));

    // when
    const result = await shareWorkbook(newClient(), "wb-7");

    // then
    expect(result).toEqual(reference);
  });

  it("should send POST request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildReference(), 201));

    // when
    await shareWorkbook(newClient(), "wb-7");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-7/share`, {
      method: "POST",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildReference(), 201));

    // when
    await shareWorkbook(newClient(), "wb/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2Fspecial%26id/share`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 409 }));

    // when / then
    await expect(shareWorkbook(newClient(), "wb-7")).rejects.toMatchObject({ status: 409 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(shareWorkbook(newClient(), "wb-7"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks referenceId", async () => {
    // given
    const { referenceId: _, ...withoutReferenceId } = buildReference();
    fetchMock.mockResolvedValue(jsonResponse(withoutReferenceId, 201));

    // when / then
    await expect(shareWorkbook(newClient(), "wb-7")).rejects.toMatchObject({ status: 502 });
  });
});

describe("unshareWorkbook", () => {
  it("should send DELETE request with correct URL and auth header", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await unshareWorkbook(newClient(), "ref-123");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/shared/ref-123`, {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode referenceId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await unshareWorkbook(newClient(), "ref/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/shared/ref%2Fspecial%26id`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 404 }));

    // when / then
    await expect(unshareWorkbook(newClient(), "ref-1")).rejects.toMatchObject({ status: 404 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(unshareWorkbook(newClient(), "ref-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});
