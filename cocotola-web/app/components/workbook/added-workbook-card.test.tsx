import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { SharedWorkbook } from "~/lib/api/sharing.server";
import { formatDate } from "~/lib/format/date";

import "~/i18n/config";

const submitMock = vi.fn();

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useNavigate: vi.fn(() => vi.fn()),
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: submitMock })),
  };
});

import { AddedWorkbookCard } from "./added-workbook-card";

function buildAdded(overrides: Partial<SharedWorkbook> = {}): SharedWorkbook {
  return {
    referenceId: "ref-1",
    workbookId: "wb-1",
    title: "TOEIC 600",
    description: "Vocabulary for TOEIC 600",
    language: "en",
    addedAt: "2026-03-01T00:00:00Z",
    ...overrides,
  };
}

function renderCard(workbook: SharedWorkbook = buildAdded()) {
  render(
    <ConfirmDialogProvider>
      <AddedWorkbookCard workbook={workbook} />
    </ConfirmDialogProvider>,
  );
}

describe("AddedWorkbookCard", () => {
  beforeEach(() => {
    submitMock.mockReset();
  });

  it("should render a study trigger", () => {
    // given / when
    renderCard();

    // then
    expect(screen.getByRole("button", { name: /Study/i })).toBeInTheDocument();
  });

  it("should render the date the workbook was added", () => {
    // given
    const addedAt = "2026-03-15T12:00:00Z";

    // when
    renderCard(buildAdded({ addedAt }));

    // then
    expect(screen.getByText(new RegExp(formatDate(addedAt, "en")))).toBeInTheDocument();
  });

  it("should not render an edit action", () => {
    // given / when
    renderCard();

    // then
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it("should submit unshare with the reference id when removal is confirmed", async () => {
    // given
    const user = userEvent.setup();
    renderCard(buildAdded({ referenceId: "ref-xyz", title: "Sample Book" }));

    // when
    await user.click(screen.getByRole("button", { name: /remove "Sample Book"/i }));
    await user.click(screen.getByRole("button", { name: "Remove from My Workbooks" }));

    // then
    expect(submitMock).toHaveBeenCalledWith(
      { intent: "unshare", referenceId: "ref-xyz" },
      { method: "post" },
    );
  });

  it("should not submit unshare when removal is canceled", async () => {
    // given
    const user = userEvent.setup();
    renderCard(buildAdded({ title: "Sample Book" }));

    // when
    await user.click(screen.getByRole("button", { name: /remove "Sample Book"/i }));
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    // then
    expect(submitMock).not.toHaveBeenCalled();
  });
});
