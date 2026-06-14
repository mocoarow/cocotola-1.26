import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/dashboard.server", () => ({
  getDashboard: vi.fn(),
}));

// Cover every concrete export the route module imports, not just the
// ones touched by the current tests. The component-only tests below do
// not exercise the loader, but any future test that does (or any
// reorganization of the route module that imports a new helper into
// component scope) would otherwise hit `undefined` instead of a mock
// when the route's import binding is evaluated.
vi.mock("~/lib/api/user-setting.server", () => ({
  getUserPreferences: vi.fn(),
  updateUserLanguage: vi.fn(),
  updateUserDailyGoal: vi.fn(),
  updateUserTimezone: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useActionData: vi.fn(),
    useNavigation: vi.fn(() => ({ state: "idle", formData: undefined })),
    Form: ({ children }: { children: ReactNode }) => <form>{children}</form>,
  };
});

import { useActionData, useLoaderData } from "react-router";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { updateUserDailyGoal, updateUserTimezone } from "~/lib/api/user-setting.server";

import DashboardPage, { action } from "./dashboard";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedUseActionData = vi.mocked(useActionData);

function makeLoaderData() {
  return {
    dashboard: {
      from: "2025-06-15",
      to: "2026-06-14",
      days: [
        { date: "2026-06-13", answeredCount: 5, correctCount: 4 },
        { date: "2026-06-14", answeredCount: 3, correctCount: 2 },
      ],
      currentStreak: 2,
      longestStreak: 7,
      todayCount: 3,
      todayCorrect: 2,
      activeDays: 2,
      totalAnswered: 8,
      totalCorrect: 6,
    },
    dailyGoal: 10,
    timezone: "Asia/Tokyo",
    language: "en",
  };
}

describe("DashboardPage", () => {
  it("should_renderHeading_when_loaderDataIsAvailable", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<DashboardPage />);

    // then
    // Role/level instead of text content: the translated dashboard title
    // changes between locales, and the i18n init order in vitest can
    // hand back an unresolved key. Asserting on the role keeps this test
    // honest about "there is exactly one <h1> on the page" regardless of
    // which language the i18n bundle chose.
    expect(screen.getByRole("heading", { level: 1 })).toBeInTheDocument();
  });

  it("should_renderStreakCard_when_dashboardHasStreakData", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("streak-card")).toBeInTheDocument();
    expect(screen.getByTestId("streak-longest").textContent).toMatch(/7/);
  });

  it("should_renderDailyGoalCard_when_dashboardHasTodayData", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("daily-goal-card")).toBeInTheDocument();
    expect(screen.getByTestId("daily-goal-progress").textContent).toMatch(/3.*10/);
  });

  it("should_renderContributionGraph_when_dashboardHasDays", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("contribution-graph")).toBeInTheDocument();
    expect(screen.getAllByTestId("contribution-cell").length).toBeGreaterThan(0);
  });

  it("should_renderPreferencesForm_when_dashboardLoads", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue(undefined);

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByLabelText(/Daily goal/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Timezone/i)).toBeInTheDocument();
  });

  it("should_showSavedConfirmation_when_actionDataIsSavedAt", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({ ok: true, savedAt: "2026-06-14T10:00:00Z" });

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("settings-saved")).toBeInTheDocument();
  });

  it("should_showPartialFailureMessage_when_actionDataReportsFailedFields", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({ ok: false, failedFields: ["timezone"] });

    // when
    render(<DashboardPage />);

    // then
    const banner = screen.getByTestId("settings-partial-failure");
    expect(banner).toBeInTheDocument();
    expect(banner.textContent).toMatch(/timezone/);
  });

  it("should_listAllFailedFields_when_bothSavesFailed", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());
    mockedUseActionData.mockReturnValue({ ok: false, failedFields: ["dailyGoal", "timezone"] });

    // when
    render(<DashboardPage />);

    // then
    const banner = screen.getByTestId("settings-partial-failure");
    expect(banner.textContent).toMatch(/dailyGoal/);
    expect(banner.textContent).toMatch(/timezone/);
  });
});

// The action carries the real validation logic and partial-failure
// reporting for the preferences form. Rendering tests above only
// confirm the UI reacts to action data — these tests exercise the
// action itself: validation branches, the sequential save semantics,
// and the redirect-propagation contract that keeps session-expiry
// from being silently swallowed.
describe("dashboard action", () => {
  function buildActionRequest(fields: Record<string, string>): Request {
    const form = new URLSearchParams(fields);
    return new Request("http://localhost/dashboard", {
      method: "POST",
      body: form.toString(),
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    });
  }

  async function callAction(request: Request) {
    // Casting around the generated Route.ActionArgs type so the tests
    // do not have to construct an empty params/context shape that the
    // action does not read.
    return action({ request } as unknown as Parameters<typeof action>[0]);
  }

  beforeEach(() => {
    vi.mocked(requireAuth).mockResolvedValue({
      accessToken: "test-token",
      refreshToken: "test-refresh",
    });
    vi.mocked(updateUserDailyGoal).mockResolvedValue(undefined);
    vi.mocked(updateUserTimezone).mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should_returnOk_when_bothUpdatesSucceed", async () => {
    // given
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    expect(result).toMatchObject({ ok: true });
    expect("savedAt" in (result as object)).toBe(true);
    expect(vi.mocked(updateUserDailyGoal)).toHaveBeenCalledWith(request, "test-token", 20);
    expect(vi.mocked(updateUserTimezone)).toHaveBeenCalledWith(request, "test-token", "Asia/Tokyo");
  });

  it("should_throw400_when_intentIsNotSavePreferences", async () => {
    // given
    const request = buildActionRequest({
      intent: "deleteEverything",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_throw400_when_dailyGoalBelowMin", async () => {
    // given
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "0",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
    // updaters must not be reached when validation fails
    expect(vi.mocked(updateUserDailyGoal)).not.toHaveBeenCalled();
  });

  it("should_throw400_when_dailyGoalHasTrailingGarbage", async () => {
    // given
    // Number.parseInt("30abc", 10) returns 30 silently. The strict
    // /^\d+$/ guard must convert this into NaN so the range check
    // rejects the tampered body rather than accepting "30" out of it.
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "30abc",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
    expect(vi.mocked(updateUserDailyGoal)).not.toHaveBeenCalled();
  });

  it("should_throw400_when_dailyGoalAboveMax", async () => {
    // given
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "501",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_throw400_when_timezoneShapeInvalid", async () => {
    // given
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "20",
      timezone: "Asia Tokyo", // space → fails character-class regex
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBeInstanceOf(Response);
    expect((caught as Response).status).toBe(400);
  });

  it("should_returnFailedTimezoneField_when_timezoneUpdaterRejects", async () => {
    // given
    vi.mocked(updateUserTimezone).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    // dailyGoal still saved (no partial silence), timezone reported as failed
    expect(result).toMatchObject({ ok: false, failedFields: ["timezone"] });
  });

  it("should_returnBothFailedFields_when_bothUpdatersReject", async () => {
    // given
    vi.mocked(updateUserDailyGoal).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    vi.mocked(updateUserTimezone).mockRejectedValue(
      new Response("server unavailable", { status: 500 }),
    );
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when
    const result = await callAction(request);

    // then
    expect(result).toMatchObject({ ok: false, failedFields: ["dailyGoal", "timezone"] });
  });

  it("should_propagateRedirect_when_updaterThrowsRedirectResponse", async () => {
    // given
    // redirectOnUnauthorized throws a redirect Response on 401. The
    // savePreferenceField helper must propagate any 3xx Response
    // (Location header present) so React Router can navigate, instead
    // of recording it as a field-level failure.
    const loginRedirect = new Response(null, {
      status: 302,
      headers: { Location: "/login" },
    });
    vi.mocked(updateUserDailyGoal).mockRejectedValue(loginRedirect);
    const request = buildActionRequest({
      intent: "savePreferences",
      dailyGoal: "20",
      timezone: "Asia/Tokyo",
    });

    // when / then
    let caught: unknown;
    try {
      await callAction(request);
    } catch (err) {
      caught = err;
    }
    expect(caught).toBe(loginRedirect);
    // timezone updater must not run after a propagated redirect
    expect(vi.mocked(updateUserTimezone)).not.toHaveBeenCalled();
  });
});
