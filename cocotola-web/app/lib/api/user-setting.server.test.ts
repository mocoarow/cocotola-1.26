import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { updateUserLanguage } from "./user-setting.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("updateUserLanguage", () => {
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
    await expect(updateUserLanguage("token", "ja")).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });

  it("should send PUT with correct URL, headers, and body on success", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await updateUserLanguage("test-token", "ja");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/user-setting/language",
      {
        method: "PUT",
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ language: "ja" }),
      },
    );
  });

  it("should resolve with no value on 204", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    const result = await updateUserLanguage("token", "en");

    // then
    expect(result).toBeUndefined();
  });

  it("should throw Response when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 400 });

    // when / then
    await expect(updateUserLanguage("token", "xx")).rejects.toBeInstanceOf(Response);
  });

  it("should throw Response when API returns 409 conflict", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 409 });

    // when / then
    await expect(updateUserLanguage("token", "ja")).rejects.toBeInstanceOf(Response);
  });
});
