import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deleteWorkbook, listWorkbooks } from "./workbook.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("listWorkbooks", () => {
  beforeEach(() => {
    vi.stubEnv("COCOTOLA_QUESTION_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should throw when COCOTOLA_QUESTION_URL is not set", async () => {
    // given
    vi.stubEnv("COCOTOLA_QUESTION_URL", "");

    // when / then
    await expect(listWorkbooks("token", "sp-1")).rejects.toThrow(
      "COCOTOLA_QUESTION_URL environment variable is required",
    );
  });

  it("should return workbooks on success", async () => {
    // given
    const workbooks = [{ workbookId: "wb-1", title: "Test Workbook", spaceId: "sp-1" }];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ workbooks }),
    });

    // when
    const result = await listWorkbooks("test-token", "sp-1");

    // then
    expect(result).toEqual(workbooks);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook?spaceId=sp-1", {
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
    const result = await listWorkbooks("token", "sp-1");

    // then
    expect(result).toEqual([]);
  });

  it("should encode spaceId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ workbooks: [] }),
    });

    // when
    await listWorkbooks("token", "space with spaces");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook?spaceId=space%20with%20spaces",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(listWorkbooks("token", "sp-1")).rejects.toBeInstanceOf(Response);
  });
});

describe("deleteWorkbook", () => {
  beforeEach(() => {
    vi.stubEnv("COCOTOLA_QUESTION_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send DELETE request with correct URL and auth header", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await deleteWorkbook("test-token", "wb-123");

    // then
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/wb-123", {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await deleteWorkbook("token", "wb/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2Fspecial%26id",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 404 });

    // when / then
    await expect(deleteWorkbook("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});
