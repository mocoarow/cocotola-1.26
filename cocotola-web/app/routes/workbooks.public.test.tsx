import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import type { PublicWorkbook, SharedReference } from "~/lib/api/sharing.server";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
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
    useFetcher: vi.fn(() => ({
      state: "idle",
      Form: ({
        children,
        onSubmit,
        ...rest
      }: {
        children: ReactNode;
        onSubmit?: (e: React.FormEvent<HTMLFormElement>) => void;
        // biome-ignore lint/suspicious/noExplicitAny: passthrough props for the mocked Form
      } & Record<string, any>) => (
        <form
          {...rest}
          onSubmit={(e) => {
            e.preventDefault();
            if (onSubmit) {
              onSubmit(e);
              return;
            }
            const formData = new FormData(e.currentTarget);
            const payload = Object.fromEntries(formData.entries());
            submitMock(payload, { method: "post" });
          }}
        >
          {children}
        </form>
      ),
      submit: submitMock,
    })),
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { useLoaderData } from "react-router";
import {
  listPublicWorkbooks,
  listSharedWorkbooks,
  shareWorkbook,
  unshareWorkbook,
} from "~/lib/api/sharing.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import WorkbooksPublic, { action, loader } from "./workbooks.public";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedRequireAuth = vi.mocked(requireAuth);
const mockedListPublic = vi.mocked(listPublicWorkbooks);
const mockedListShared = vi.mocked(listSharedWorkbooks);
const mockedShareWorkbook = vi.mocked(shareWorkbook);
const mockedUnshareWorkbook = vi.mocked(unshareWorkbook);

function buildWorkbook(overrides: Partial<PublicWorkbook> = {}): PublicWorkbook {
  return {
    workbookId: "wb-1",
    ownerId: "user-1",
    title: "English Vocabulary",
    description: "Basic English words",
    language: "en",
    createdAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function buildReference(overrides: Partial<SharedReference> = {}): SharedReference {
  return {
    referenceId: "ref-1",
    workbookId: "wb-1",
    addedAt: "2026-04-10T00:00:00Z",
    ...overrides,
  };
}

function renderPublic() {
  return render(
    <ConfirmDialogProvider>
      <WorkbooksPublic />
    </ConfirmDialogProvider>,
  );
}

describe("WorkbooksPublic component", () => {
  beforeEach(() => {
    submitMock.mockReset();
  });

  it("should render empty state when no public workbooks exist", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbooks: [], references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByText("No public workbooks")).toBeInTheDocument();
    expect(
      screen.getByText("There are no public workbooks available in your language yet."),
    ).toBeInTheDocument();
  });

  it("should render page title and description", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbooks: [], references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByText("Public Workbooks")).toBeInTheDocument();
    expect(
      screen.getByText(
        "Browse public workbooks shared by other users and add them to your library.",
      ),
    ).toBeInTheDocument();
  });

  it("should render public workbook cards when workbooks exist", () => {
    // given
    const workbooks = [
      buildWorkbook({ workbookId: "wb-1", title: "English Vocabulary" }),
      buildWorkbook({ workbookId: "wb-2", title: "French Grammar", language: "fr" }),
    ];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByText("English Vocabulary")).toBeInTheDocument();
    expect(screen.getByText("French Grammar")).toBeInTheDocument();
  });

  it("should render the language badge for each workbook", () => {
    // given
    const workbooks = [buildWorkbook({ language: "en" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByText("English")).toBeInTheDocument();
  });

  it("should render description on workbook card", () => {
    // given
    const workbooks = [buildWorkbook({ description: "Learn basic words" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByText("Learn basic words")).toBeInTheDocument();
  });

  it("should render Add button when workbook is not yet shared", () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });

    // when
    renderPublic();

    // then
    expect(screen.getByRole("button", { name: /add "English Vocabulary"/i })).toBeInTheDocument();
    expect(screen.queryByText("Added")).not.toBeInTheDocument();
  });

  it("should render Study link even when workbook is not yet shared", () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });

    // when
    renderPublic();

    // then
    const studyLink = screen.getByText("Study").closest("a");
    expect(studyLink).toBeInTheDocument();
    expect(studyLink).toHaveAttribute("href", "/workbooks/wb-1/study");
  });

  it("should render Added button and Study link when workbook is already shared", () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1" })];
    const references = [buildReference({ workbookId: "wb-1", referenceId: "ref-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references });

    // when
    renderPublic();

    // then
    expect(screen.getByText("Added")).toBeInTheDocument();
    const studyLink = screen.getByText("Study").closest("a");
    expect(studyLink).toBeInTheDocument();
    expect(studyLink).toHaveAttribute("href", "/workbooks/wb-1/study");
  });

  it("should submit share intent when Add button is clicked", async () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-add" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references: [] });
    const user = userEvent.setup();
    renderPublic();

    // when
    await user.click(screen.getByRole("button", { name: /add "English Vocabulary"/i }));

    // then
    expect(submitMock).toHaveBeenCalledWith(
      { intent: "share", workbookId: "wb-add" },
      { method: "post" },
    );
  });

  it("should open confirm dialog and submit unshare intent when confirmed", async () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1", title: "Sample Book" })];
    const references = [buildReference({ workbookId: "wb-1", referenceId: "ref-xyz" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references });
    const user = userEvent.setup();
    renderPublic();

    // when
    await user.click(screen.getByRole("button", { name: /remove "Sample Book"/i }));
    expect(screen.getByText('Remove "Sample Book" from your library?')).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Remove from library" }));

    // then
    expect(submitMock).toHaveBeenCalledWith(
      { intent: "unshare", referenceId: "ref-xyz" },
      { method: "post" },
    );
  });

  it("should not submit unshare when confirm dialog is canceled", async () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1", title: "Sample Book" })];
    const references = [buildReference({ workbookId: "wb-1", referenceId: "ref-xyz" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references });
    const user = userEvent.setup();
    renderPublic();

    // when
    await user.click(screen.getByRole("button", { name: /remove "Sample Book"/i }));
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    // then
    expect(submitMock).not.toHaveBeenCalled();
  });

  it("should match references by workbookId across multiple workbooks", () => {
    // given
    const workbooks = [
      buildWorkbook({ workbookId: "wb-1", title: "Shared Book" }),
      buildWorkbook({ workbookId: "wb-2", title: "Unshared Book" }),
    ];
    const references = [buildReference({ workbookId: "wb-1", referenceId: "ref-1" })];
    mockedUseLoaderData.mockReturnValue({ workbooks, references });

    // when
    renderPublic();

    // then
    expect(screen.getByText("Added")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /add "Unshared Book"/i })).toBeInTheDocument();
  });
});

type LoaderArgs = Parameters<typeof loader>[0];
type ActionArgs = Parameters<typeof action>[0];

function buildLoaderArgs(request: Request): LoaderArgs {
  return { request, params: {}, context: {} } as unknown as LoaderArgs;
}

function buildActionArgs(request: Request): ActionArgs {
  return { request, params: {}, context: {} } as unknown as ActionArgs;
}

describe("WorkbooksPublic loader", () => {
  beforeEach(() => {
    mockedRequireAuth.mockResolvedValue({ accessToken: "test-token", refreshToken: undefined });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it("should fetch public workbooks and shared references in parallel", async () => {
    // given
    const workbooks = [buildWorkbook({ workbookId: "wb-1" })];
    const references = [buildReference({ workbookId: "wb-1" })];
    mockedListPublic.mockResolvedValue(workbooks);
    mockedListShared.mockResolvedValue(references);

    // when
    const result = await loader(buildLoaderArgs(new Request("http://localhost/workbooks/public")));

    // then
    expect(result).toEqual({ workbooks, references });
    expect(mockedListPublic).toHaveBeenCalledWith("test-token");
    expect(mockedListShared).toHaveBeenCalledWith("test-token");
  });

  it("should propagate errors when listPublicWorkbooks fails", async () => {
    // given
    mockedListPublic.mockRejectedValue(new Response("boom", { status: 500 }));
    mockedListShared.mockResolvedValue([]);

    // when / then
    await expect(
      loader(buildLoaderArgs(new Request("http://localhost/workbooks/public"))),
    ).rejects.toBeInstanceOf(Response);
  });
});

describe("WorkbooksPublic action", () => {
  beforeEach(() => {
    mockedRequireAuth.mockResolvedValue({ accessToken: "test-token", refreshToken: undefined });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  function buildPostRequest(formEntries: Record<string, string>): Request {
    const body = new URLSearchParams(formEntries).toString();
    return new Request("http://localhost/workbooks/public", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body,
    });
  }

  it("should call shareWorkbook on share intent", async () => {
    // given
    mockedShareWorkbook.mockResolvedValue({
      referenceId: "ref-new",
      workbookId: "wb-1",
      addedAt: "2026-04-29T00:00:00Z",
    });
    const request = buildPostRequest({ intent: "share", workbookId: "wb-1" });

    // when
    const result = await action(buildActionArgs(request));

    // then
    expect(mockedShareWorkbook).toHaveBeenCalledWith("test-token", "wb-1");
    expect(result).toEqual({ ok: true, alreadyShared: false });
  });

  it("should tolerate 409 conflict on share intent (already shared)", async () => {
    // given
    mockedShareWorkbook.mockRejectedValue(new Response("conflict", { status: 409 }));
    const request = buildPostRequest({ intent: "share", workbookId: "wb-1" });

    // when
    const result = await action(buildActionArgs(request));

    // then
    expect(result).toEqual({ ok: true, alreadyShared: true });
  });

  it("should propagate non-409 errors on share intent", async () => {
    // given
    mockedShareWorkbook.mockRejectedValue(new Response("boom", { status: 500 }));
    const request = buildPostRequest({ intent: "share", workbookId: "wb-1" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
  });

  it("should reject share intent without workbookId", async () => {
    // given
    const request = buildPostRequest({ intent: "share" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
    expect(mockedShareWorkbook).not.toHaveBeenCalled();
  });

  it("should reject share intent with empty workbookId", async () => {
    // given
    const request = buildPostRequest({ intent: "share", workbookId: "" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
    expect(mockedShareWorkbook).not.toHaveBeenCalled();
  });

  it("should call unshareWorkbook on unshare intent", async () => {
    // given
    mockedUnshareWorkbook.mockResolvedValue(undefined);
    const request = buildPostRequest({ intent: "unshare", referenceId: "ref-1" });

    // when
    const result = await action(buildActionArgs(request));

    // then
    expect(mockedUnshareWorkbook).toHaveBeenCalledWith("test-token", "ref-1");
    expect(result).toEqual({ ok: true, alreadyRemoved: false });
  });

  it("should tolerate 404 on unshare intent (already removed)", async () => {
    // given
    mockedUnshareWorkbook.mockRejectedValue(new Response("not found", { status: 404 }));
    const request = buildPostRequest({ intent: "unshare", referenceId: "ref-1" });

    // when
    const result = await action(buildActionArgs(request));

    // then
    expect(result).toEqual({ ok: true, alreadyRemoved: true });
  });

  it("should propagate non-404 errors on unshare intent", async () => {
    // given
    mockedUnshareWorkbook.mockRejectedValue(new Response("boom", { status: 500 }));
    const request = buildPostRequest({ intent: "unshare", referenceId: "ref-1" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
  });

  it("should reject unshare intent without referenceId", async () => {
    // given
    const request = buildPostRequest({ intent: "unshare" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
    expect(mockedUnshareWorkbook).not.toHaveBeenCalled();
  });

  it("should reject unknown intent", async () => {
    // given
    const request = buildPostRequest({ intent: "bogus" });

    // when / then
    await expect(action(buildActionArgs(request))).rejects.toBeInstanceOf(Response);
  });
});
