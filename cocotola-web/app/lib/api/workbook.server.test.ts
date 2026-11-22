import { format } from "node:util";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { catchError, jsonResponse, newClient } from "./test-helpers";
import {
  createWorkbook,
  deleteWorkbook,
  getWorkbook,
  listWorkbooks,
  updateWorkbook,
  type Workbook,
  WorkbookValidationError,
} from "./workbook.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";

function unauthorized(): Response {
  return new Response(null, { status: 401 });
}

function buildWorkbook(overrides: Partial<Workbook> = {}): Workbook {
  return {
    workbookId: "wb-1",
    spaceId: "sp-1",
    ownerId: "user-1",
    organizationId: "org-1",
    title: "Phrasal verbs",
    description: "Common phrasal verbs",
    visibility: "public",
    language: "fr",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-04-10T00:00:00Z",
    ...overrides,
  };
}

const CREATE_DATA = {
  spaceId: "sp-1",
  title: "New Workbook",
  description: "desc",
  visibility: "private" as const,
  language: "fr",
};

const UPDATE_DATA = {
  title: "Updated Title",
  description: "desc",
  visibility: "private" as const,
  language: "fr",
};

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

describe("listWorkbooks", () => {
  it("should throw when QUESTION_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("QUESTION_BASE_URL", "");

    // when / then
    await expect(listWorkbooks(newClient(), "sp-1")).rejects.toThrow(
      "QUESTION_BASE_URL environment variable is required",
    );
  });

  it("should return workbooks on success", async () => {
    // given
    const workbooks = [buildWorkbook()];
    fetchMock.mockResolvedValue(jsonResponse({ workbooks }));

    // when
    const result = await listWorkbooks(newClient(), "sp-1");

    // then
    expect(result).toEqual(workbooks);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [] }));

    // when
    await listWorkbooks(newClient(), "sp-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook?spaceId=sp-1`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when workbooks is undefined", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const result = await listWorkbooks(newClient(), "sp-1");

    // then
    expect(result).toEqual([]);
  });

  it("should return empty array when workbooks is null", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: null }));

    // when
    const result = await listWorkbooks(newClient(), "sp-1");

    // then
    expect(result).toEqual([]);
  });

  it("should encode spaceId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [] }));

    // when
    await listWorkbooks(newClient(), "space with spaces");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook?spaceId=space%20with%20spaces`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(listWorkbooks(newClient(), "sp-1")).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorized());

    // when
    const caught = await catchError(listWorkbooks(newClient(), "sp-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a workbook in the response is malformed", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [{ workbookId: "wb-1" }] }));

    // when / then
    await expect(listWorkbooks(newClient(), "sp-1")).rejects.toMatchObject({ status: 502 });
  });
});

describe("createWorkbook", () => {
  it("should return the created workbook on success", async () => {
    // given
    const workbook = buildWorkbook({ workbookId: "wb-new" });
    fetchMock.mockResolvedValue(jsonResponse(workbook, 201));

    // when
    const result = await createWorkbook(newClient(), CREATE_DATA);

    // then
    expect(result).toEqual(workbook);
  });

  it("should send POST request with correct body and headers", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildWorkbook(), 201));

    // when
    await createWorkbook(newClient(), CREATE_DATA);

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook`, {
      method: "POST",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(CREATE_DATA),
    });
  });

  it("should not log the title via console.info", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildWorkbook(), 201));
    const title = "secret-title-7f3a";

    // when
    await createWorkbook(newClient(), { ...CREATE_DATA, title });

    // then
    const logged = vi
      .mocked(console.info)
      .mock.calls.map((args) => format(...args))
      .join("\n");
    expect(logged).not.toContain(title);
  });

  it("should throw WorkbookValidationError when API rejects the input", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));

    // when / then
    await expect(createWorkbook(newClient(), CREATE_DATA)).rejects.toBeInstanceOf(
      WorkbookValidationError,
    );
  });

  it("should throw Response when API returns a server error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(createWorkbook(newClient(), CREATE_DATA)).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorized());

    // when
    const caught = await catchError(createWorkbook(newClient(), CREATE_DATA));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response has an unknown visibility", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ...buildWorkbook(), visibility: "secret" }, 201));

    // when / then
    await expect(createWorkbook(newClient(), CREATE_DATA)).rejects.toMatchObject({ status: 502 });
  });
});

describe("deleteWorkbook", () => {
  it("should send DELETE request with correct URL and auth header", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteWorkbook(newClient(), "wb-123");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-123`, {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteWorkbook(newClient(), "wb/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2Fspecial%26id`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 404 }));

    // when / then
    await expect(deleteWorkbook(newClient(), "wb-1")).rejects.toMatchObject({ status: 404 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorized());

    // when
    const caught = await catchError(deleteWorkbook(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});

describe("getWorkbook", () => {
  it("should return workbook on success", async () => {
    // given
    const workbook = buildWorkbook();
    fetchMock.mockResolvedValue(jsonResponse(workbook));

    // when
    const result = await getWorkbook(newClient(), "wb-1");

    // then
    expect(result).toEqual(workbook);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildWorkbook()));

    // when
    await getWorkbook(newClient(), "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-1`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildWorkbook()));

    // when
    await getWorkbook(newClient(), "wb/special&id");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2Fspecial%26id`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 404 }));

    // when / then
    await expect(getWorkbook(newClient(), "wb-1")).rejects.toMatchObject({ status: 404 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorized());

    // when
    const caught = await catchError(getWorkbook(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks ownerId", async () => {
    // given
    const { ownerId: _, ...withoutOwner } = buildWorkbook();
    fetchMock.mockResolvedValue(jsonResponse(withoutOwner));

    // when / then
    await expect(getWorkbook(newClient(), "wb-1")).rejects.toMatchObject({ status: 502 });
  });
});

describe("updateWorkbook", () => {
  it("should return the updated workbook on success", async () => {
    // given
    const workbook = buildWorkbook({ title: "Updated Title" });
    fetchMock.mockResolvedValue(jsonResponse(workbook));

    // when
    const result = await updateWorkbook(newClient(), "wb-1", UPDATE_DATA);

    // then
    expect(result).toEqual(workbook);
  });

  it("should send PUT request with correct body and headers", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildWorkbook()));

    // when
    await updateWorkbook(newClient(), "wb-1", UPDATE_DATA);

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-1`, {
      method: "PUT",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(UPDATE_DATA),
    });
  });

  it("should throw WorkbookValidationError when API rejects the input", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));

    // when / then
    await expect(updateWorkbook(newClient(), "wb-1", UPDATE_DATA)).rejects.toBeInstanceOf(
      WorkbookValidationError,
    );
  });

  it("should throw Response when API returns a server error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(updateWorkbook(newClient(), "wb-1", UPDATE_DATA)).rejects.toMatchObject({
      status: 500,
    });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorized());

    // when
    const caught = await catchError(updateWorkbook(newClient(), "wb-1", UPDATE_DATA));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response is not a workbook", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ok: true }));

    // when / then
    await expect(updateWorkbook(newClient(), "wb-1", UPDATE_DATA)).rejects.toMatchObject({
      status: 502,
    });
  });
});
