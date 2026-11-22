import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("~/lib/supabase/server", () => ({
  createSupabaseServerClient: vi.fn(),
}));

vi.mock("~/lib/api/auth.server", () => ({
  exchangeSupabaseToken: vi.fn(),
  refreshAccessToken: vi.fn(),
}));

vi.mock("~/lib/api/client.server", () => ({
  createBackendClient: vi.fn(),
}));

vi.mock("~/lib/api/user-setting.server", () => ({
  getUserPreferences: vi.fn(),
  getUserPreferencesOrNull: vi.fn(),
  updateUserLanguage: vi.fn(),
  updateUserDailyGoal: vi.fn(),
  updateUserTimezone: vi.fn(),
}));

import { exchangeSupabaseToken } from "~/lib/api/auth.server";
import { createBackendClient } from "~/lib/api/client.server";
import { backendClientSentinel } from "~/lib/api/test-helpers";
import { getUserPreferences, updateUserLanguage } from "~/lib/api/user-setting.server";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import { loader } from "./auth.callback";

type LoaderArgs = Parameters<typeof loader>[0];

const backendClient = backendClientSentinel();

const EXCHANGED_ACCESS_TOKEN = "exchanged-access-token";

const preferences = {
  userId: "user-1",
  loginId: "learner-login",
  organizationName: "org-name",
  language: "en",
  dailyGoal: 25,
  timezone: "America/Los_Angeles",
};

function givenSupabaseSession(): void {
  const supabase = {
    auth: {
      exchangeCodeForSession: vi.fn().mockResolvedValue({
        data: { session: { user: { id: "supabase-user-1" }, access_token: "supabase-token" } },
        error: null,
      }),
    },
  };
  vi.mocked(createSupabaseServerClient).mockReturnValue(
    supabase as unknown as ReturnType<typeof createSupabaseServerClient>,
  );
}

function buildLoaderArgs(): LoaderArgs {
  const request = new Request("http://localhost/auth/callback?code=auth-code", {
    headers: { "Accept-Language": "ja" },
  });
  return { request, params: {}, context: {} } as unknown as LoaderArgs;
}

describe("AuthCallback loader", () => {
  beforeEach(() => {
    vi.spyOn(console, "info").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    givenSupabaseSession();
    vi.mocked(exchangeSupabaseToken).mockResolvedValue({
      accessToken: EXCHANGED_ACCESS_TOKEN,
      refreshToken: "exchanged-refresh-token",
    });
    vi.mocked(createBackendClient).mockReturnValue(backendClient);
    vi.mocked(getUserPreferences).mockResolvedValue(preferences);
    vi.mocked(updateUserLanguage).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.resetAllMocks();
  });

  it("should create the backend client with the request the loader received", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].request).toBe(args.request);
  });

  it("should create the backend client with the exchanged access token", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].accessToken).toBe(
      EXCHANGED_ACCESS_TOKEN,
    );
  });

  it("should fetch the user preferences through the backend client", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(getUserPreferences).mock.calls[0][0]).toBe(backendClient);
  });

  it("should sync the language through the backend client", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(updateUserLanguage).mock.calls[0][0]).toBe(backendClient);
  });

  it("should sync the UI language when the persisted language is the backend default", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(updateUserLanguage).mock.calls[0][1]).toBe("ja");
  });

  it("should rethrow the /login redirect when the preferences lookup throws it", async () => {
    // given
    const loginRedirect = new Response(null, { status: 302, headers: { Location: "/login" } });
    vi.mocked(getUserPreferences).mockRejectedValue(loginRedirect);

    // when
    const caught = await loader(buildLoaderArgs()).catch((e: unknown) => e);

    // then
    expect(caught).toBe(loginRedirect);
  });

  it("should redirect to / when the preferences lookup fails with 502", async () => {
    // given
    vi.mocked(getUserPreferences).mockRejectedValue(new Response("bad", { status: 502 }));

    // when
    const result = await loader(buildLoaderArgs());

    // then
    expect((result as Response).headers.get("Location")).toBe("/");
  });

  it("should keep the session cookie when the preferences lookup fails with 502", async () => {
    // given
    vi.mocked(getUserPreferences).mockRejectedValue(new Response("bad", { status: 502 }));

    // when
    const result = await loader(buildLoaderArgs());

    // then
    expect((result as Response).headers.get("Set-Cookie")).toMatch(/__cocotola_session=/);
  });

  it("should redirect to / when the preferences lookup fails with a non-Response error", async () => {
    // given
    vi.mocked(getUserPreferences).mockRejectedValue(new TypeError("network"));

    // when
    const result = await loader(buildLoaderArgs());

    // then
    expect((result as Response).headers.get("Location")).toBe("/");
  });
});
