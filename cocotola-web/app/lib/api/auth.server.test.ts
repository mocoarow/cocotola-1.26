import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { refreshAccessToken } from "./auth.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("refreshAccessToken", () => {
  beforeEach(() => {
    vi.stubEnv("AUTH_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
    fetchMock.mockReset();
  });

  it("should POST the refresh token to the refresh endpoint", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(200, { accessToken: "new-at" }));

    // when
    await refreshAccessToken("rt");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/refresh",
      expect.objectContaining({
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken: "rt" }),
      }),
    );
  });

  it("should pass an abort signal to fetch", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(200, { accessToken: "new-at" }));

    // when
    await refreshAccessToken("rt");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });

  it("should time out the refresh request after 3 seconds", async () => {
    // given
    const timeoutSpy = vi.spyOn(AbortSignal, "timeout");
    fetchMock.mockResolvedValue(jsonResponse(200, { accessToken: "new-at" }));

    // when
    await refreshAccessToken("rt");

    // then
    expect(timeoutSpy).toHaveBeenCalledWith(3_000);
  });

  it("should return the new access token when the API returns 200", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(200, { accessToken: "new-at" }));

    // when
    const result = await refreshAccessToken("rt");

    // then
    expect(result).toEqual({ kind: "refreshed", accessToken: "new-at" });
  });

  it("should return unauthorized when the API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(401, { code: "invalid_refresh_token" }));

    // when
    const result = await refreshAccessToken("rt");

    // then
    expect(result).toEqual({ kind: "unauthorized" });
  });

  it("should reject when the API returns 500", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(500, { code: "internal_server_error" }));

    // when / then
    await expect(refreshAccessToken("rt")).rejects.toThrow("Token refresh failed (500)");
  });

  it.each([
    { name: "empty", body: { accessToken: "" } },
    { name: "missing", body: {} },
    { name: "a number", body: { accessToken: 123 } },
    { name: "in a null body", body: null },
  ])("should reject when accessToken in the 200 response is $name", async ({ body }) => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(200, body));

    // when / then
    await expect(refreshAccessToken("rt")).rejects.toThrow(
      "Token refresh returned no access token",
    );
  });

  it("should reject when fetch times out", async () => {
    // given
    fetchMock.mockRejectedValue(new DOMException("timed out", "TimeoutError"));

    // when / then
    await expect(refreshAccessToken("rt")).rejects.toThrow("timed out");
  });

  it("should reject when AUTH_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("AUTH_BASE_URL", "");

    // when / then
    await expect(refreshAccessToken("rt")).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });
});
