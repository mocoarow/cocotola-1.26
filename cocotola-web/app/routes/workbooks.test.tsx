import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/hooks/use-mobile", () => ({
  useIsMobile: vi.fn(() => false),
}));

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
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

import type { UserPreferences } from "~/lib/api/user-setting.server";
import { ErrorBoundary } from "./workbooks";

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
