import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import "~/i18n/config";
import type { DashboardDailyItem } from "~/lib/api/dashboard.server";

import { ContributionGraph } from "./contribution-graph";

function makeItems(rows: { date: string; count: number }[]): DashboardDailyItem[] {
  return rows.map((r) => ({ date: r.date, answeredCount: r.count, correctCount: r.count }));
}

describe("ContributionGraph", () => {
  it("should_renderEmptyState_when_itemsIsEmpty", () => {
    // given / when
    render(<ContributionGraph items={[]} from="2026-06-01" to="2026-06-07" locale="en-US" />);

    // then
    expect(screen.getByTestId("contribution-graph-empty")).toBeInTheDocument();
  });

  it("should_renderOneCellPerDay_when_itemsProvided", () => {
    // given
    const items = makeItems([
      { date: "2026-06-01", count: 0 },
      { date: "2026-06-02", count: 1 },
      { date: "2026-06-03", count: 5 },
      { date: "2026-06-04", count: 10 },
      { date: "2026-06-05", count: 20 },
      { date: "2026-06-06", count: 3 },
      { date: "2026-06-07", count: 0 },
    ]);

    // when
    render(<ContributionGraph items={items} from="2026-06-01" to="2026-06-07" locale="en-US" />);

    // then
    const cells = screen.getAllByTestId("contribution-cell");
    expect(cells).toHaveLength(7);
  });

  it("should_mapAnsweredCountToFiveTones_when_buildingCells", () => {
    // given
    const items = makeItems([
      { date: "2026-06-01", count: 0 },
      { date: "2026-06-02", count: 1 },
      { date: "2026-06-03", count: 5 },
      { date: "2026-06-04", count: 10 },
      { date: "2026-06-05", count: 20 },
    ]);

    // when
    render(<ContributionGraph items={items} from="2026-06-01" to="2026-06-05" locale="en-US" />);

    // then
    const cells = screen.getAllByTestId("contribution-cell");
    expect(cells[0]).toHaveAttribute("data-tone", "0");
    expect(cells[1]).toHaveAttribute("data-tone", "1");
    expect(cells[2]).toHaveAttribute("data-tone", "2");
    expect(cells[3]).toHaveAttribute("data-tone", "3");
    expect(cells[4]).toHaveAttribute("data-tone", "4");
  });

  it("should_includeDateAndCountInCellTitle_when_renderingTooltip", () => {
    // given
    const items = makeItems([{ date: "2026-06-14", count: 7 }]);

    // when
    render(<ContributionGraph items={items} from="2026-06-14" to="2026-06-14" locale="en-US" />);

    // then
    const graph = screen.getByTestId("contribution-graph");
    const cell = within(graph).getByTestId("contribution-cell");
    expect(cell.getAttribute("title")).toContain("2026-06-14");
    expect(cell.getAttribute("title")).toMatch(/7/);
  });
});
