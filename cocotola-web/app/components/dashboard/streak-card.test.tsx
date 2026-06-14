import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import "~/i18n/config";

import { StreakCard } from "./streak-card";

describe("StreakCard", () => {
  it("should_renderCurrentStreak_when_called", () => {
    // given / when
    render(<StreakCard current={5} longest={9} />);

    // then
    expect(screen.getByTestId("streak-card")).toBeInTheDocument();
    expect(screen.getByText(/5/)).toBeInTheDocument();
  });

  it("should_renderLongestStreak_when_called", () => {
    // given / when
    render(<StreakCard current={1} longest={42} />);

    // then
    expect(screen.getByTestId("streak-longest").textContent).toMatch(/42/);
  });

  it("should_renderZeroValues_when_userHasNoActivity", () => {
    // given / when
    render(<StreakCard current={0} longest={0} />);

    // then
    expect(screen.getByTestId("streak-card")).toBeInTheDocument();
  });
});
