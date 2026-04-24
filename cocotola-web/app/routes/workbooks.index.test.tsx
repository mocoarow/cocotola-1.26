import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
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

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useFetcher: vi.fn(() => ({ state: "idle", Form: "form" })),
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
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("English Vocabulary")).toBeInTheDocument();
    expect(screen.getByText("Math Problems")).toBeInTheDocument();
  });

  it("should render visibility badge on workbook card", () => {
    // given
    const workbooks = [createWorkbook({ visibility: "public" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("public")).toBeInTheDocument();
  });

  it("should render description on workbook card", () => {
    // given
    const workbooks = [createWorkbook({ description: "Learn basic words" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("Learn basic words")).toBeInTheDocument();
  });

  it("should render Study button as disabled and Edit button as link", () => {
    // given
    const workbooks = [createWorkbook({ workbookId: "wb-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(<WorkbooksIndex />);

    // then
    expect(screen.getByText("Study").closest("button")).toBeDisabled();
    const editLink = screen.getByText("Edit").closest("a");
    expect(editLink).toBeInTheDocument();
    expect(editLink).toHaveAttribute("href", "/workbooks/wb-1");
  });

  it("should render delete button with hidden workbookId input", () => {
    // given
    const workbooks = [createWorkbook({ workbookId: "wb-123" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, spaceId: "sp-1" });

    // when
    render(<WorkbooksIndex />);

    // then
    const hiddenInput = document.querySelector('input[name="workbookId"][value="wb-123"]');
    expect(hiddenInput).toBeInTheDocument();
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
