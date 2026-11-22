import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/hooks/use-mobile", () => ({
  useIsMobile: vi.fn(() => false),
}));

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/client.server", () => ({
  createBackendClient: vi.fn(),
}));

vi.mock("~/lib/api/user-setting.server", () => ({
  getUserPreferences: vi.fn(),
  getUserPreferencesOrNull: vi.fn(),
  updateUserLanguage: vi.fn(),
  updateUserDailyGoal: vi.fn(),
  updateUserTimezone: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLocation: vi.fn(() => ({ pathname: "/workbooks/wb-1" })),
    Form: ({ children }: { children: ReactNode }) => <form>{children}</form>,
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { createBackendClient } from "~/lib/api/client.server";
import { backendClientSentinel } from "~/lib/api/test-helpers";
import { getUserPreferencesOrNull, type UserPreferences } from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { ErrorBoundary, loader } from "./workbooks";

type ErrorBoundaryProps = Parameters<typeof ErrorBoundary>[0];

function renderBoundary(status: number, loaderData: { user: UserPreferences | null } | undefined) {
  const props = {
    error: { status, statusText: "", internal: false, data: "" },
    params: {},
    loaderData,
  } as unknown as ErrorBoundaryProps;
  return render(<ErrorBoundary {...props} />);
}

describe("WorkbooksLayout ErrorBoundary", () => {
  it("should keep the sidebar when a child route fails", () => {
    // given
    const loaderData = { user: null };

    // when
    renderBoundary(409, loaderData);

    // then
    expect(screen.getByTestId("workbooks-sidebar")).toBeInTheDocument();
  });

  it("should render the error view inside the layout when a child route fails", () => {
    // given
    const loaderData = { user: null };

    // when
    renderBoundary(409, loaderData);

    // then
    expect(screen.getByTestId("route-error-conflict")).toBeInTheDocument();
  });

  it("should render the signed-in user when the layout loader succeeded", () => {
    // given
    const user = {
      userId: "u-1",
      loginId: "learner-login",
      organizationName: "org-name",
    } as UserPreferences;

    // when
    renderBoundary(500, { user });

    // then
    expect(screen.getByTestId("workbooks-user-login-id")).toHaveTextContent("learner-login");
  });

  it("should keep the sidebar when the layout loader data is unavailable", () => {
    // given
    const loaderData = undefined;

    // when
    renderBoundary(500, loaderData);

    // then
    expect(screen.getByTestId("workbooks-sidebar")).toBeInTheDocument();
  });
});

type LoaderArgs = Parameters<typeof loader>[0];

const backendClient = backendClientSentinel();

describe("WorkbooksLayout loader", () => {
  beforeEach(() => {
    vi.spyOn(console, "info").mockImplementation(() => {});
    vi.mocked(requireAuth).mockReturnValue({ accessToken: "test-token" });
    vi.mocked(createBackendClient).mockReturnValue(backendClient);
    vi.mocked(getUserPreferencesOrNull).mockResolvedValue(null);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.clearAllMocks();
  });

  function buildLoaderArgs(): LoaderArgs {
    const request = new Request("http://localhost/workbooks");
    return { request, params: {}, context: {} } as unknown as LoaderArgs;
  }

  it("should create the backend client with the request the loader received", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].request).toBe(args.request);
  });

  it("should create the backend client with the access token", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(createBackendClient).mock.calls[0][0].accessToken).toBe("test-token");
  });

  it("should fetch the user preferences through the backend client", async () => {
    // given
    const args = buildLoaderArgs();

    // when
    await loader(args);

    // then
    expect(vi.mocked(getUserPreferencesOrNull).mock.calls[0][0]).toBe(backendClient);
  });
});
