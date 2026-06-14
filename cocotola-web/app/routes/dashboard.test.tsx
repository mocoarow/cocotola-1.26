import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/dashboard.server", () => ({
  getDashboard: vi.fn(),
}));

// Cover every export of user-setting.server even though dashboard.tsx
// currently only imports getUserPreferences. Defensive completeness:
// when a future change pulls another helper from this module into route
// scope (or a future test exercises the loader through a different
// path), the import binding evaluates against a real mock instead of
// `undefined`. This mirrors the settings.test.tsx pattern so neither
// test starts to drift independently when the module grows.
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
  };
});

import { useLoaderData } from "react-router";

import DashboardPage from "./dashboard";

const mockedUseLoaderData = vi.mocked(useLoaderData);

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
  };
}

describe("DashboardPage", () => {
  it("should_renderHeading_when_loaderDataIsAvailable", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());

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

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("streak-card")).toBeInTheDocument();
    expect(screen.getByTestId("streak-longest").textContent).toMatch(/7/);
  });

  it("should_renderDailyGoalCard_when_dashboardHasTodayData", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("daily-goal-card")).toBeInTheDocument();
    expect(screen.getByTestId("daily-goal-progress").textContent).toMatch(/3.*10/);
  });

  it("should_renderContributionGraph_when_dashboardHasDays", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());

    // when
    render(<DashboardPage />);

    // then
    expect(screen.getByTestId("contribution-graph")).toBeInTheDocument();
    expect(screen.getAllByTestId("contribution-cell").length).toBeGreaterThan(0);
  });

  it("should_notRenderPreferencesForm_after_settingsPageMovedOut", () => {
    // given
    mockedUseLoaderData.mockReturnValue(makeLoaderData());

    // when
    render(<DashboardPage />);

    // then: preferences inputs now live on /settings, not the dashboard
    expect(screen.queryByLabelText(/Daily goal/i)).not.toBeInTheDocument();
    expect(screen.queryByLabelText(/Timezone/i)).not.toBeInTheDocument();
  });
});
