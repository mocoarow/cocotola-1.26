import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

import type { StudySummary } from "~/lib/api/study.server";

const navigateMock = vi.fn();
const fetcherLoadMock = vi.fn();
type FetcherState = "idle" | "loading" | "submitting";
let fetcherState: { state: FetcherState; data: StudySummary | undefined };

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useNavigate: () => navigateMock,
    useFetcher: () => ({
      state: fetcherState.state,
      data: fetcherState.data,
      load: fetcherLoadMock,
    }),
  };
});

import { StartStudyDialog } from "./start-study-dialog";

function makeSummary(totalDue: number, overrides: Partial<StudySummary> = {}): StudySummary {
  return {
    newCount: Math.max(0, totalDue - 1),
    reviewCount: totalDue > 0 ? 1 : 0,
    totalDue,
    reviewRatioNumerator: 1,
    reviewRatioDenominator: 4,
    ...overrides,
  };
}

function renderDialog() {
  return render(<StartStudyDialog workbookId="wb-1" triggerLabel="Study" />);
}

async function openDialog() {
  const user = userEvent.setup();
  renderDialog();
  await user.click(screen.getByRole("button", { name: "Study" }));
  return user;
}

beforeEach(() => {
  navigateMock.mockReset();
  fetcherLoadMock.mockReset();
  fetcherState = { state: "idle", data: undefined };
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("StartStudyDialog", () => {
  describe("data loading", () => {
    it("should request summary the first time the dialog is opened", async () => {
      // given
      fetcherState = { state: "idle", data: undefined };

      // when
      await openDialog();

      // then
      expect(fetcherLoadMock).toHaveBeenCalledWith("/workbooks/wb-1/study-summary");
    });

    it("should show loading text while the summary is being fetched", async () => {
      // given
      fetcherState = { state: "loading", data: undefined };

      // when
      await openDialog();

      // then
      expect(screen.getByText("Loading available questions...")).toBeInTheDocument();
    });

    it("should not refetch when reopened while data is already cached", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(30) };
      const user = userEvent.setup();
      renderDialog();

      // when
      await user.click(screen.getByRole("button", { name: "Study" }));

      // then
      expect(fetcherLoadMock).not.toHaveBeenCalled();
    });
  });

  describe("size options", () => {
    it("should show only the 'All' option when totalDue is below the step", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(5) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      const optionValues = Array.from(select.options).map((o) => Number(o.value));
      expect(optionValues).toEqual([5]);
      expect(within(select).getByText("All (5)")).toBeInTheDocument();
    });

    it("should not duplicate the 'All' option when totalDue is exactly a step multiple", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(20) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      const optionValues = Array.from(select.options).map((o) => Number(o.value));
      expect(optionValues).toEqual([10, 20]);
    });

    it("should append an 'All' option when totalDue is between steps", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(25) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      const optionValues = Array.from(select.options).map((o) => Number(o.value));
      expect(optionValues).toEqual([10, 20, 25]);
      expect(within(select).getByText("All (25)")).toBeInTheDocument();
    });

    it("should cap options at ABSOLUTE_MAX when totalDue exceeds it", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(150) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      const optionValues = Array.from(select.options).map((o) => Number(o.value));
      expect(optionValues).toEqual([10, 20, 30, 40, 50, 60, 70, 80, 90, 100]);
    });

    it("should default the selection to 20 when available", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(50) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      expect(Number(select.value)).toBe(20);
    });

    it("should fall back to the largest option when default is not available", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(15) };

      // when
      await openDialog();

      // then
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      // sizeOptions = [10, 15], default 20 is missing → falls back to 15
      expect(Number(select.value)).toBe(15);
    });
  });

  describe("empty state", () => {
    it("should show 'nothing to study' when totalDue is zero", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(0) };

      // when
      await openDialog();

      // then
      expect(screen.getByText("Nothing to study right now.")).toBeInTheDocument();
      expect(screen.queryByLabelText("Number of questions")).not.toBeInTheDocument();
    });

    it("should disable the start button when nothing is available", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(0) };

      // when
      await openDialog();

      // then
      const startButton = screen.getByRole("button", { name: "Start" });
      expect(startButton).toBeDisabled();
    });
  });

  describe("starting a session", () => {
    it("should navigate with the selected limit when start is clicked", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(50) };
      const user = await openDialog();

      // when
      await user.click(screen.getByRole("button", { name: "Start" }));

      // then
      expect(navigateMock).toHaveBeenCalledWith("/workbooks/wb-1/study?limit=20");
    });

    it("should navigate with the user-selected limit after change", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(50) };
      const user = await openDialog();

      // when
      await user.selectOptions(screen.getByLabelText("Number of questions"), "40");
      await user.click(screen.getByRole("button", { name: "Start" }));

      // then
      expect(navigateMock).toHaveBeenCalledWith("/workbooks/wb-1/study?limit=40");
    });
  });

  describe("invalid select values", () => {
    it("should keep prior selection when onChange receives a non-finite value", async () => {
      // given
      fetcherState = { state: "idle", data: makeSummary(50) };
      const user = await openDialog();
      const select = screen.getByLabelText("Number of questions") as HTMLSelectElement;
      await user.selectOptions(select, "30");

      // when: inject a synthetic change event whose value parses to NaN
      const nanOption = document.createElement("option");
      nanOption.value = "not-a-number";
      select.appendChild(nanOption);
      select.value = "not-a-number";
      select.dispatchEvent(new Event("change", { bubbles: true }));

      // then: pressing start still navigates with the last valid selection (30)
      await user.click(screen.getByRole("button", { name: "Start" }));
      expect(navigateMock).toHaveBeenCalledWith("/workbooks/wb-1/study?limit=30");
    });
  });
});
