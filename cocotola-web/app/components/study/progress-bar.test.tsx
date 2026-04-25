import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import "~/i18n/config";

import { ProgressBar } from "./progress-bar";

describe("ProgressBar", () => {
  it("should render current and total", () => {
    // given / when
    render(<ProgressBar current={2} total={10} />);

    // then
    expect(screen.getByText("3 / 10")).toBeInTheDocument();
    expect(screen.getByText("20%")).toBeInTheDocument();
  });

  it("should render 0% when total is 0", () => {
    // given / when
    render(<ProgressBar current={0} total={0} />);

    // then
    expect(screen.getByText("0%")).toBeInTheDocument();
  });

  it("should render progress bar with correct width", () => {
    // given / when
    const { container } = render(<ProgressBar current={5} total={10} />);

    // then
    const bar = container.querySelector("[style]");
    expect(bar).toHaveStyle({ width: "50%" });
  });
});
