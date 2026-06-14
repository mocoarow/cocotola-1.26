import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getUserPreferences, updateUserLanguage } from "./user-setting.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

function fakeRequest(): Request {
  return new Request("http://localhost/some-loader", { method: "GET" });
}

describe("updateUserLanguage", () => {
  beforeEach(() => {
    vi.stubEnv("AUTH_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
    fetchMock.mockReset();
  });

  it("should throw when AUTH_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("AUTH_BASE_URL", "");

    // when / then
    await expect(updateUserLanguage(fakeRequest(), "token", "ja")).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });

  it("should send PUT with correct URL, headers, and body on success", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true, status: 204 });

    // when
    await updateUserLanguage(fakeRequest(), "test-token", "ja");

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
    fetchMock.mockResolvedValue({ ok: true, status: 204 });

    // when
    const result = await updateUserLanguage(fakeRequest(), "token", "en");

    // then
    expect(result).toBeUndefined();
  });

  it("should throw Response when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 400 });

    // when / then
    await expect(updateUserLanguage(fakeRequest(), "token", "xx")).rejects.toBeInstanceOf(Response);
  });

  it("should throw Response when API returns 409 conflict", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 409 });

    // when / then
    await expect(updateUserLanguage(fakeRequest(), "token", "ja")).rejects.toBeInstanceOf(Response);
  });

  it("should throw redirect Response to /login when API returns 401", async () => {
    // given
    // redirectOnUnauthorized funnels 401 into destroySession + throw
    // redirect("/login"), so the thrown Response carries a 302 Location
    // header pointing at /login rather than the original 401.
    fetchMock.mockResolvedValue({ ok: false, status: 401 });

    // when / then
    let caught: unknown;
    try {
      await updateUserLanguage(fakeRequest(), "token", "ja");
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    const resp = caught as Response;
    expect(resp.status).toBe(302);
    expect(resp.headers.get("Location")).toBe("/login");
    expect(resp.headers.get("Set-Cookie")).toMatch(/__cocotola_session/);
  });
});

// The parser lives behind getUserPreferences so it is exercised through
// fetch mocking rather than imported directly — the function itself is
// intentionally not exported. These cases pin the type-safety contract
// at the HTTP boundary: a backend response that drifts shape (wrong
// type, missing field, NaN, non-object) must not crash the loader or
// leak stringly-typed numbers downstream.
describe("getUserPreferences (response parser)", () => {
  beforeEach(() => {
    vi.stubEnv("AUTH_BASE_URL", "http://localhost:8080");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
    fetchMock.mockReset();
  });

  function mockJsonResponse(body: unknown): void {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => body,
    });
  }

  it("should_parseAllFields_when_backendReturnsValidShape", async () => {
    // given
    mockJsonResponse({
      userId: "11111111-1111-1111-1111-111111111111",
      loginId: "alice@example.com",
      organizationName: "Acme",
      language: "ja",
      dailyGoal: 25,
      timezone: "America/Los_Angeles",
    });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs).toEqual({
      userId: "11111111-1111-1111-1111-111111111111",
      loginId: "alice@example.com",
      organizationName: "Acme",
      language: "ja",
      dailyGoal: 25,
      timezone: "America/Los_Angeles",
    });
  });

  it("should_applyAllDefaults_when_backendReturnsEmptyObject", async () => {
    // given
    mockJsonResponse({});

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    // Defaults mirror cocotola-auth/domain/user_setting.go; this test
    // pins the single source of truth on the TypeScript side.
    expect(prefs).toEqual({
      userId: "",
      loginId: "",
      organizationName: "",
      language: "en",
      dailyGoal: 10,
      timezone: "Asia/Tokyo",
    });
  });

  it("should_applyDailyGoalDefault_when_dailyGoalIsString", async () => {
    // given
    // Backend regression: returning "10" (string) instead of 10. The
    // raw `as Partial<UserPreferences>` cast would have happily passed
    // a string through. The typed parser must reject and fall back.
    mockJsonResponse({ dailyGoal: "10" });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.dailyGoal).toBe(10);
    expect(typeof prefs.dailyGoal).toBe("number");
  });

  it("should_applyDailyGoalDefault_when_dailyGoalIsNaN", async () => {
    // given
    mockJsonResponse({ dailyGoal: Number.NaN });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_applyDailyGoalDefault_when_dailyGoalIsInfinity", async () => {
    // given
    mockJsonResponse({ dailyGoal: Number.POSITIVE_INFINITY });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_applyStringDefault_when_languageIsNumber", async () => {
    // given
    mockJsonResponse({ language: 42 });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.language).toBe("en");
  });

  it("should_applyAllDefaults_when_responseIsNull", async () => {
    // given
    mockJsonResponse(null);

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.userId).toBe("");
    expect(prefs.dailyGoal).toBe(10);
    expect(prefs.timezone).toBe("Asia/Tokyo");
  });

  it("should_applyAllDefaults_when_responseIsNonObjectPrimitive", async () => {
    // given
    // A misbehaving backend returning a bare string instead of an
    // object must not crash the loader.
    mockJsonResponse("oops");

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.userId).toBe("");
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_keepValidFieldsAndDefaultInvalid_when_mixedShape", async () => {
    // given
    mockJsonResponse({
      userId: "good",
      dailyGoal: "bad",
      timezone: "Asia/Tokyo",
      language: null,
    });

    // when
    const prefs = await getUserPreferences(fakeRequest(), "token");

    // then
    expect(prefs.userId).toBe("good");          // valid → kept
    expect(prefs.dailyGoal).toBe(10);            // string → default
    expect(prefs.timezone).toBe("Asia/Tokyo");  // valid → kept
    expect(prefs.language).toBe("en");           // null → default
  });
});
