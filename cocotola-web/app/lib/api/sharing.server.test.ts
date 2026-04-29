import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  listPublicWorkbooks,
  listSharedWorkbooks,
  shareWorkbook,
  unshareWorkbook,
} from "./sharing.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("listPublicWorkbooks", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should throw when QUESTION_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("QUESTION_BASE_URL", "");

    // when / then
    await expect(listPublicWorkbooks("token")).rejects.toThrow(
      "QUESTION_BASE_URL environment variable is required",
    );
  });

  it("should return public workbooks on success", async () => {
    // given
    const workbooks = [
      {
        workbookId: "wb-1",
        ownerId: "user-1",
        title: "English Vocabulary",
        description: "Basic words",
        language: "en",
        createdAt: "2026-01-01T00:00:00Z",
      },
    ];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ workbooks }),
    });

    // when
    const result = await listPublicWorkbooks("test-token");

    // then
    expect(result).toEqual(workbooks);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/public", {
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when workbooks is undefined", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    const result = await listPublicWorkbooks("token");

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(listPublicWorkbooks("token")).rejects.toBeInstanceOf(Response);
  });
});

describe("listSharedWorkbooks", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return shared references on success", async () => {
    // given
    const references = [
      {
        referenceId: "ref-1",
        workbookId: "wb-1",
        addedAt: "2026-04-10T00:00:00Z",
      },
    ];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ references }),
    });

    // when
    const result = await listSharedWorkbooks("test-token");

    // then
    expect(result).toEqual(references);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/shared", {
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when references is undefined", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    const result = await listSharedWorkbooks("token");

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 403 });

    // when / then
    await expect(listSharedWorkbooks("token")).rejects.toBeInstanceOf(Response);
  });
});

describe("shareWorkbook", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request and return the new reference", async () => {
    // given
    const reference = {
      referenceId: "ref-new",
      workbookId: "wb-1",
      addedAt: "2026-04-29T00:00:00Z",
    };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(reference),
    });

    // when
    const result = await shareWorkbook("test-token", "wb-1");

    // then
    expect(result).toEqual(reference);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/wb-1/share", {
      method: "POST",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ referenceId: "r", workbookId: "x", addedAt: "" }),
    });

    // when
    await shareWorkbook("token", "wb/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2Fspecial%26id/share",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 409 });

    // when / then
    await expect(shareWorkbook("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});

describe("unshareWorkbook", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send DELETE request with correct URL and auth header", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await unshareWorkbook("test-token", "ref-123");

    // then
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/shared/ref-123", {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode referenceId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await unshareWorkbook("token", "ref/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/shared/ref%2Fspecial%26id",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 404 });

    // when / then
    await expect(unshareWorkbook("token", "ref-1")).rejects.toBeInstanceOf(Response);
  });
});
