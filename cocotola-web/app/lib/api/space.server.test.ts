import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { findPrivateSpace, listSpaces, type Space } from "./space.server";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const AUTH_BASE_URL = "http://localhost:8080";

function buildSpace(overrides: Partial<Space> = {}): Space {
  return {
    spaceId: "sp-1",
    organizationId: "org-1",
    ownerId: "user-1",
    keyName: "private@@user1",
    name: "Private(user1)",
    spaceType: "private",
    deleted: false,
    ...overrides,
  };
}

beforeEach(() => {
  vi.stubEnv("AUTH_BASE_URL", AUTH_BASE_URL);
  vi.spyOn(console, "info").mockImplementation(() => {});
  vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
  fetchMock.mockReset();
});

describe("listSpaces", () => {
  it("should throw when AUTH_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("AUTH_BASE_URL", "");

    // when / then
    await expect(listSpaces(newClient())).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });

  it("should return spaces on success", async () => {
    // given
    const spaces = [buildSpace()];
    fetchMock.mockResolvedValue(jsonResponse({ spaces }));

    // when
    const result = await listSpaces(newClient());

    // then
    expect(result).toEqual(spaces);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ spaces: [] }));

    // when
    await listSpaces(newClient());

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${AUTH_BASE_URL}/api/v1/auth/space`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when spaces is undefined", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const result = await listSpaces(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should return empty array when spaces is null", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ spaces: null }));

    // when
    const result = await listSpaces(newClient());

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 403 }));

    // when / then
    await expect(listSpaces(newClient())).rejects.toMatchObject({ status: 403 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(listSpaces(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a space has an unknown spaceType", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ spaces: [{ ...buildSpace(), spaceType: "team" }] }));

    // when / then
    await expect(listSpaces(newClient())).rejects.toMatchObject({ status: 502 });
  });
});

describe("findPrivateSpace", () => {
  it("should return the private non-deleted space", async () => {
    // given
    const spaces = [
      buildSpace({ spaceId: "sp-pub", spaceType: "public", keyName: "public@@org" }),
      buildSpace({ spaceId: "sp-priv" }),
    ];
    fetchMock.mockResolvedValue(jsonResponse({ spaces }));

    // when
    const result = await findPrivateSpace(newClient());

    // then
    expect(result?.spaceId).toBe("sp-priv");
  });

  it("should return undefined when no private space exists", async () => {
    // given
    fetchMock.mockResolvedValue(
      jsonResponse({ spaces: [buildSpace({ spaceId: "sp-pub", spaceType: "public" })] }),
    );

    // when
    const result = await findPrivateSpace(newClient());

    // then
    expect(result).toBeUndefined();
  });

  it("should skip deleted private spaces", async () => {
    // given
    fetchMock.mockResolvedValue(
      jsonResponse({ spaces: [buildSpace({ spaceId: "sp-del", deleted: true })] }),
    );

    // when
    const result = await findPrivateSpace(newClient());

    // then
    expect(result).toBeUndefined();
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(findPrivateSpace(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the spaces in the response are malformed", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ spaces: [{ spaceId: "sp-priv" }] }));

    // when / then
    await expect(findPrivateSpace(newClient())).rejects.toMatchObject({ status: 502 });
  });
});
