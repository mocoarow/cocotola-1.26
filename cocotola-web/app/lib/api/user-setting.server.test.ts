import { format } from "node:util";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";
import {
  getUserPreferences,
  getUserPreferencesOrNull,
  updateUserDailyGoal,
  updateUserLanguage,
  updateUserTimezone,
} from "./user-setting.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const AUTH_BASE_URL = "http://localhost:8080";

function noContentResponse(): Response {
  return new Response(null, { status: 204 });
}

function spyConsoleOutput(): () => string {
  const spies = (["log", "info", "warn", "error", "debug"] as const).map((level) =>
    vi.spyOn(console, level).mockImplementation(() => {}),
  );
  return () =>
    spies
      .flatMap((spy) => spy.mock.calls)
      .map((args) => format(...args))
      .join("\n");
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

describe("updateUserLanguage", () => {
  it("should throw when AUTH_BASE_URL is not set", async () => {
    // given
    vi.stubEnv("AUTH_BASE_URL", "");

    // when / then
    await expect(updateUserLanguage(newClient(), "ja")).rejects.toThrow(
      "AUTH_BASE_URL environment variable is required",
    );
  });

  it("should send PUT with correct URL, headers, and body on success", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());

    // when
    await updateUserLanguage(newClient(), "ja");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${AUTH_BASE_URL}/api/v1/auth/user-setting/language`, {
      method: "PUT",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ language: "ja" }),
    });
  });

  it("should resolve with no value on 204", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());

    // when
    const result = await updateUserLanguage(newClient(), "en");

    // then
    expect(result).toBeUndefined();
  });

  it("should throw Response when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));

    // when / then
    await expect(updateUserLanguage(newClient(), "xx")).rejects.toMatchObject({ status: 400 });
  });

  it("should throw Response when API returns 409 conflict", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 409 }));

    // when / then
    await expect(updateUserLanguage(newClient(), "ja")).rejects.toMatchObject({ status: 409 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(updateUserLanguage(newClient(), "ja"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should not log the language value", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());
    const consoleOutput = spyConsoleOutput();

    // when
    await updateUserLanguage(newClient(), "sentinel-language");

    // then
    expect(consoleOutput()).not.toContain("sentinel-language");
  });
});

describe("updateUserDailyGoal", () => {
  it("should send PUT with correct URL, headers, and body on success", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());

    // when
    await updateUserDailyGoal(newClient(), 25);

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${AUTH_BASE_URL}/api/v1/auth/user-setting/daily-goal`, {
      method: "PUT",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ dailyGoal: 25 }),
    });
  });

  it("should throw Response when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));

    // when / then
    await expect(updateUserDailyGoal(newClient(), 25)).rejects.toMatchObject({ status: 400 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(updateUserDailyGoal(newClient(), 25));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});

describe("updateUserTimezone", () => {
  it("should send PUT with correct URL, headers, and body on success", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());

    // when
    await updateUserTimezone(newClient(), "America/Los_Angeles");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${AUTH_BASE_URL}/api/v1/auth/user-setting/timezone`, {
      method: "PUT",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ timezone: "America/Los_Angeles" }),
    });
  });

  it("should throw Response when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));

    // when / then
    await expect(updateUserTimezone(newClient(), "Asia/Tokyo")).rejects.toMatchObject({
      status: 400,
    });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(updateUserTimezone(newClient(), "Asia/Tokyo"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should not log the timezone value", async () => {
    // given
    fetchMock.mockResolvedValue(noContentResponse());
    const consoleOutput = spyConsoleOutput();

    // when
    await updateUserTimezone(newClient(), "Sentinel/Timezone");

    // then
    expect(consoleOutput()).not.toContain("Sentinel/Timezone");
  });
});

describe("getUserPreferences", () => {
  it("should send GET request to /auth/me with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    await getUserPreferences(newClient());

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${AUTH_BASE_URL}/api/v1/auth/me`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(getUserPreferences(newClient())).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(getUserPreferences(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});

// These cases pin the type-safety contract at the HTTP boundary: a backend
// response that drifts shape must fall back to defaults instead of failing.
describe("getUserPreferences (response parser)", () => {
  it("should_parseAllFields_when_backendReturnsValidShape", async () => {
    // given
    fetchMock.mockResolvedValue(
      jsonResponse({
        userId: "11111111-1111-1111-1111-111111111111",
        loginId: "alice@example.com",
        organizationName: "Acme",
        language: "ja",
        dailyGoal: 25,
        timezone: "America/Los_Angeles",
      }),
    );

    // when
    const prefs = await getUserPreferences(newClient());

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
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const prefs = await getUserPreferences(newClient());

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
    fetchMock.mockResolvedValue(jsonResponse({ dailyGoal: "10" }));

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_applyDailyGoalDefault_when_dailyGoalIsNull", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ dailyGoal: null }));

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_applyDailyGoalDefault_when_dailyGoalOverflowsToInfinity", async () => {
    // given
    fetchMock.mockResolvedValue(
      new Response('{"dailyGoal":1e400}', {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs.dailyGoal).toBe(10);
  });

  it("should_applyStringDefault_when_languageIsNumber", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ language: 42 }));

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs.language).toBe("en");
  });

  it("should_applyAllDefaults_when_responseIsNull", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(null));

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs).toEqual({
      userId: "",
      loginId: "",
      organizationName: "",
      language: "en",
      dailyGoal: 10,
      timezone: "Asia/Tokyo",
    });
  });

  it("should_applyAllDefaults_when_responseIsNonObjectPrimitive", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse("oops"));

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs).toEqual({
      userId: "",
      loginId: "",
      organizationName: "",
      language: "en",
      dailyGoal: 10,
      timezone: "Asia/Tokyo",
    });
  });

  it("should_keepValidFieldsAndDefaultInvalid_when_mixedShape", async () => {
    // given
    fetchMock.mockResolvedValue(
      jsonResponse({
        userId: "good",
        dailyGoal: "bad",
        timezone: "America/Los_Angeles",
        language: null,
      }),
    );

    // when
    const prefs = await getUserPreferences(newClient());

    // then
    expect(prefs).toEqual({
      userId: "good",
      loginId: "",
      organizationName: "",
      language: "en",
      dailyGoal: 10,
      timezone: "America/Los_Angeles",
    });
  });
});

describe("getUserPreferencesOrNull", () => {
  it("should return the preferences when /auth/me succeeds", async () => {
    // given
    fetchMock.mockResolvedValue(
      jsonResponse({ userId: "user-1", timezone: "America/Los_Angeles" }),
    );

    // when
    const prefs = await getUserPreferencesOrNull(newClient());

    // then
    expect(prefs?.userId).toBe("user-1");
  });

  it("should return null when /auth/me returns 503", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 503 }));

    // when
    const prefs = await getUserPreferencesOrNull(newClient());

    // then
    expect(prefs).toBeNull();
  });

  it("should return null when the /auth/me body is not JSON", async () => {
    // given
    // The client turns an unreadable body into a 502 Response.
    fetchMock.mockResolvedValue(new Response("<html>", { status: 200 }));

    // when
    const prefs = await getUserPreferencesOrNull(newClient());

    // then
    expect(prefs).toBeNull();
  });

  it("should return null when fetch rejects", async () => {
    // given
    fetchMock.mockRejectedValue(new TypeError("network"));

    // when
    const prefs = await getUserPreferencesOrNull(newClient());

    // then
    expect(prefs).toBeNull();
  });

  it("should throw redirect to /login when /auth/me returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(getUserPreferencesOrNull(newClient()));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});
