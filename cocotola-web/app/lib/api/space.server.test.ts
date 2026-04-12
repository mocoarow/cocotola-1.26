import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { findPrivateSpace, listSpaces } from "./space.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("listSpaces", () => {
  beforeEach(() => {
    vi.stubEnv("AUTH_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should throw when AUTH_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("AUTH_BASE_URL", "");

    // when / then
    await expect(listSpaces("token")).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });

  it("should return spaces on success", async () => {
    // given
    const spaces = [
      {
        spaceId: "sp-1",
        spaceType: "private",
        deleted: false,
        keyName: "private@@user1",
        name: "Private(user1)",
        organizationId: "org-1",
        ownerId: "user-1",
      },
    ];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ spaces }),
    });

    // when
    const result = await listSpaces("test-token");

    // then
    expect(result).toEqual(spaces);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8080/api/v1/auth/space", {
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when spaces is undefined", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    const result = await listSpaces("test-token");

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 403 });

    // when / then
    await expect(listSpaces("token")).rejects.toBeInstanceOf(Response);
  });
});

describe("findPrivateSpace", () => {
  beforeEach(() => {
    vi.stubEnv("AUTH_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return the private non-deleted space", async () => {
    // given
    const spaces = [
      {
        spaceId: "sp-pub",
        spaceType: "public" as const,
        deleted: false,
        keyName: "public@@org",
        name: "Public",
        organizationId: "org-1",
        ownerId: "user-1",
      },
      {
        spaceId: "sp-priv",
        spaceType: "private" as const,
        deleted: false,
        keyName: "private@@user1",
        name: "Private",
        organizationId: "org-1",
        ownerId: "user-1",
      },
    ];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ spaces }),
    });

    // when
    const result = await findPrivateSpace("token");

    // then
    expect(result?.spaceId).toBe("sp-priv");
  });

  it("should return undefined when no private space exists", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve({ spaces: [{ spaceId: "sp-pub", spaceType: "public", deleted: false }] }),
    });

    // when
    const result = await findPrivateSpace("token");

    // then
    expect(result).toBeUndefined();
  });

  it("should skip deleted private spaces", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve({ spaces: [{ spaceId: "sp-del", spaceType: "private", deleted: true }] }),
    });

    // when
    const result = await findPrivateSpace("token");

    // then
    expect(result).toBeUndefined();
  });
});
