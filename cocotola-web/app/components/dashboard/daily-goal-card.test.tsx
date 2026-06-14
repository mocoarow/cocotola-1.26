import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import "~/i18n/config";

import { DailyGoalCard } from "./daily-goal-card";

describe("DailyGoalCard", () => {
  it("should_renderProgressXOfY_when_partialProgress", () => {
    // given / when
    render(<DailyGoalCard todayCount={3} dailyGoal={10} />);

    // then
    expect(screen.getByTestId("daily-goal-progress").textContent).toMatch(/3.*10/);
    expect(screen.getByTestId("daily-goal-progress").getAttribute("data-reached")).toBe("false");
  });

  it("should_markReached_when_todayCountMeetsGoal", () => {
    // given / when
    render(<DailyGoalCard todayCount={10} dailyGoal={10} />);

    // then
    expect(screen.getByTestId("daily-goal-progress").getAttribute("data-reached")).toBe("true");
  });

  it("should_capProgressBar_when_todayExceedsGoal", () => {
    // given / when
    render(<DailyGoalCard todayCount={20} dailyGoal={10} />);

    // then
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "100");
  });

  it("should_renderZeroPercent_when_dailyGoalIsZero", () => {
    // given / when
    render(<DailyGoalCard todayCount={5} dailyGoal={0} />);

    // then
    expect(screen.getByRole("progressbar")).toHaveAttribute("aria-valuenow", "0");
  });
});
