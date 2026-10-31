import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { SharedWorkbook } from "~/lib/api/sharing.server";
import type { Workbook } from "~/lib/api/workbook.server";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/space.server", () => ({
  listSpaces: vi.fn(),
  findPrivateSpace: vi.fn(),
}));

vi.mock("~/lib/api/workbook.server", () => ({
  listWorkbooks: vi.fn(),
  createWorkbook: vi.fn(),
  getWorkbook: vi.fn(),
  updateWorkbook: vi.fn(),
  deleteWorkbook: vi.fn(),
}));

vi.mock("~/lib/api/sharing.server", () => ({
  listPublicWorkbooks: vi.fn(),
  listSharedWorkbooks: vi.fn(),
  shareWorkbook: vi.fn(),
  unshareWorkbook: vi.fn(),
}));

const submitMock = vi.fn();

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useSearchParams: vi.fn(() => [new URLSearchParams(), vi.fn()]),
    useNavigate: vi.fn(() => vi.fn()),
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form", submit: submitMock })),
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { useLoaderData, useSearchParams } from "react-router";
import { listSharedWorkbooks, unshareWorkbook } from "~/lib/api/sharing.server";
import { findPrivateSpace } from "~/lib/api/space.server";
import { listWorkbooks } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import WorkbooksIndex, { action, loader } from "./workbooks.index";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedUseSearchParams = vi.mocked(useSearchParams);

function buildAdded(overrides: Partial<SharedWorkbook> = {}): SharedWorkbook {
  return {
    referenceId: "ref-1",
    workbookId: "wb-added-1",
    title: "TOEIC 600",
    description: "Vocabulary for TOEIC 600",
    language: "en",
    addedAt: "2026-03-01T00:00:00Z",
    ...overrides,
  };
}

function withTab(tab: string) {
  mockedUseSearchParams.mockReturnValue([new URLSearchParams({ tab }), vi.fn()]);
}

beforeEach(() => {
  mockedUseSearchParams.mockReturnValue([new URLSearchParams(), vi.fn()]);
});

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
    mockedUseLoaderData.mockReturnValue({ workbooks: [], addedWorkbooks: [], spaceId: null });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });

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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });
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
    mockedUseLoaderData.mockReturnValue({ workbooks, addedWorkbooks: [], spaceId: "sp-1" });
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
    mockedUseLoaderData.mockReturnValue({ workbooks: [], addedWorkbooks: [], spaceId: null });

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
    mockedUseLoaderData.mockReturnValue({ workbooks: [], addedWorkbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("New Workbook")).toBeInTheDocument();
  });

  it("should mark the tab from the query string as current", () => {
    // given
    withTab("added");
    mockedUseLoaderData.mockReturnValue({ workbooks: [], addedWorkbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByRole("link", { current: "page" })).toHaveAttribute("href", "?tab=added");
  });

  it("should link to public workbooks when the added tab is empty", () => {
    // given
    withTab("added");
    mockedUseLoaderData.mockReturnValue({ workbooks: [], addedWorkbooks: [], spaceId: null });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByRole("link", { name: "Browse public workbooks" })).toHaveAttribute(
      "href",
      "/workbooks/public",
    );
  });

  it("should render added workbook cards on the added tab", () => {
    // given
    withTab("added");
    mockedUseLoaderData.mockReturnValue({
      workbooks: [createWorkbook({ title: "My Own Book" })],
      addedWorkbooks: [buildAdded({ title: "TOEIC 600" })],
      spaceId: "sp-1",
    });

    // when
    render(
      <ConfirmDialogProvider>
        <WorkbooksIndex />
      </ConfirmDialogProvider>,
    );

    // then
    expect(screen.getAllByRole("heading", { level: 3 }).map((h) => h.textContent)).toEqual([
      "TOEIC 600",
    ]);
  });
});

type LoaderArgs = Parameters<typeof loader>[0];
type ActionArgs = Parameters<typeof action>[0];

describe("WorkbooksIndex loader", () => {
  beforeEach(() => {
    vi.mocked(requireAuth).mockResolvedValue({
      accessToken: "test-token",
      refreshToken: undefined,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should return created and added workbooks when the private space exists", async () => {
    // given
    const created = [createWorkbook({ workbookId: "wb-9" })];
    const added = [buildAdded()];
    vi.mocked(findPrivateSpace).mockResolvedValue({
      spaceId: "sp-9",
      organizationId: "org-1",
      ownerId: "user-1",
      keyName: "private@@user-1",
      name: "Private",
      spaceType: "private",
      deleted: false,
    });
    vi.mocked(listWorkbooks).mockImplementation(async (token, spaceId) =>
      token === "test-token" && spaceId === "sp-9" ? created : [],
    );
    vi.mocked(listSharedWorkbooks).mockResolvedValue(added);
    const args = { request: new Request("http://localhost/workbooks") } as unknown as LoaderArgs;

    // when
    const result = await loader(args);

    // then
    expect(result).toEqual({ workbooks: created, addedWorkbooks: added, spaceId: "sp-9" });
  });

  it("should return added workbooks when the private space does not exist", async () => {
    // given
    const added = [buildAdded()];
    vi.mocked(findPrivateSpace).mockResolvedValue(undefined);
    vi.mocked(listSharedWorkbooks).mockResolvedValue(added);
    const args = { request: new Request("http://localhost/workbooks") } as unknown as LoaderArgs;

    // when
    const result = await loader(args);

    // then
    expect(result).toEqual({ workbooks: [], addedWorkbooks: added, spaceId: null });
  });
});

describe("WorkbooksIndex action", () => {
  beforeEach(() => {
    vi.mocked(requireAuth).mockResolvedValue({
      accessToken: "test-token",
      refreshToken: undefined,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  function buildActionArgs(formEntries: Record<string, string>): ActionArgs {
    const request = new Request("http://localhost/workbooks", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams(formEntries).toString(),
    });
    return { request, params: {}, context: {} } as unknown as ActionArgs;
  }

  it("should call unshareWorkbook with the reference id on unshare intent", async () => {
    // given
    vi.mocked(unshareWorkbook).mockResolvedValue(undefined);

    // when
    await action(buildActionArgs({ intent: "unshare", referenceId: "ref-xyz" }));

    // then
    expect(unshareWorkbook).toHaveBeenCalledWith("test-token", "ref-xyz");
  });

  it("should reject unshare intent without referenceId", async () => {
    // given
    const args = buildActionArgs({ intent: "unshare" });

    // when / then
    await expect(action(args)).rejects.toMatchObject({ status: 400 });
  });
});
