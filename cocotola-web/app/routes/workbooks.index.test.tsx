import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { Workbook } from "~/lib/api/workbook.server";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/space.server", () => ({
  findPrivateSpace: vi.fn(),
}));

vi.mock("~/lib/api/workbook.server", () => ({
  listWorkbooks: vi.fn(),
  deleteWorkbook: vi.fn(),
  createWorkbook: vi.fn(),
}));

const submitMock = vi.fn();

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useNavigate: vi.fn(() => vi.fn()),
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: submitMock })),
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { useLoaderData } from "react-router";
import WorkbooksIndex from "./workbooks.index";

const mockedUseLoaderData = vi.mocked(useLoaderData);

function createWorkbook(overrides: Partial<Workbook> = {}): Workbook {
  return {
    workbookId: "wb-1",
    spaceId: "sp-1",
    ownerId: "user-1",
    organizationId: "org-1",
    title: "English Vocabulary",
    description: "Basic English words",
    visibility: "private",
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-04-10T00:00:00Z",
    ...overrides,
  };
}

describe("WorkbooksIndex", () => {
  it("should render empty state when no workbooks", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("No workbooks yet")).toBeInTheDocument();
    expect(screen.getByText("Create your first workbook to get started.")).toBeInTheDocument();
  });

  it("should render workbook cards when workbooks exist", () => {
    // given
    const workbooks = [
      createWorkbook({ workbookId: "wb-1", title: "English Vocabulary" }),
      createWorkbook({ workbookId: "wb-2", title: "Math Problems", visibility: "public" }),
    ];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    expect(screen.getByText("English Vocabulary")).toBeInTheDocument();
    expect(screen.getByText("Math Problems")).toBeInTheDocument();
  });

  it("should render visibility badge on workbook card", () => {
    // given
    const workbooks = [createWorkbook({ visibility: "public" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    expect(screen.getByText("public")).toBeInTheDocument();
  });

  it("should render description on workbook card", () => {
    // given
    const workbooks = [createWorkbook({ description: "Learn basic words" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    expect(screen.getByText("Learn basic words")).toBeInTheDocument();
  });

  it("should render Study as a button (opens picker dialog) and Edit as link", () => {
    // given: the Study trigger now opens a question-count picker dialog,
    // so it must be a <button> rather than an <a>. The actual loader
    // navigation happens after the user confirms in the dialog.
    const workbooks = [createWorkbook({ workbookId: "wb-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    const studyTrigger = screen.getByRole("button", { name: /Study/i });
    expect(studyTrigger).toBeInTheDocument();
    expect(studyTrigger.tagName).toBe("BUTTON");
    const editLink = screen.getByText("Edit").closest("a");
    expect(editLink).toBeInTheDocument();
    expect(editLink).toHaveAttribute("href", "/workbooks/wb-1");
  });

  it("should render delete button", () => {
    // given
    const workbooks = [createWorkbook({ workbookId: "wb-123" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    expect(screen.getByText("Delete")).toBeInTheDocument();
  });

  it("should open confirm dialog and submit delete when confirmed", async () => {
    // given
    submitMock.mockReset();
    const workbooks = [createWorkbook({ workbookId: "wb-123", title: "Sample Book" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // when
    const deleteTrigger = screen.getByText("Delete").closest("button");
    expect(deleteTrigger).not.toBeNull();
    await user.click(deleteTrigger as HTMLElement);
    expect(screen.getByText('Delete "Sample Book"?')).toBeInTheDocument();
    const confirmButtons = screen.getAllByRole("button").filter((b) => b.textContent === "Delete");
    expect(confirmButtons.length).toBeGreaterThan(0);
    await user.click(confirmButtons[confirmButtons.length - 1]);

    // then
    expect(submitMock).toHaveBeenCalledWith(
      { intent: "delete", workbookId: "wb-123" },
      { method: "post" },
    );
  });

  it("should not submit delete when confirm dialog is canceled", async () => {
    // given
    submitMock.mockReset();
    const workbooks = [createWorkbook({ workbookId: "wb-456", title: "Other Book" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });
    const user = userEvent.setup();
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // when
    const deleteTrigger = screen.getByText("Delete").closest("button");
    expect(deleteTrigger).not.toBeNull();
    await user.click(deleteTrigger as HTMLElement);
    await user.click(screen.getByText("Cancel"));

    // then
    expect(submitMock).not.toHaveBeenCalled();
  });

  it("should show page title and description", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("My Workbooks")).toBeInTheDocument();
    expect(
      screen.getByText("Select a workbook to study or manage its problems."),
    ).toBeInTheDocument();
  });

  it("should render New Workbook button", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("New Workbook")).toBeInTheDocument();
  });
});
