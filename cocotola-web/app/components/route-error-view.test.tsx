import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { RouteErrorView } from "./route-error-view";

function routeErrorResponse(status: number) {
  return { status, statusText: "", internal: false, data: "" };
}

describe("RouteErrorView", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it.each([
    [403, "route-error-forbidden"],
    [404, "route-error-notFound"],
    [409, "route-error-conflict"],
    [500, "route-error-server"],
    [400, "route-error-unexpected"],
  ])("should render the view for status %i", (status, testId) => {
    // given / when
    render(<RouteErrorView error={routeErrorResponse(status)} />);

    // then
    expect(screen.getByTestId(testId)).toBeInTheDocument();
  });

  it.each([
    403, 404, 409, 500, 400,
  ])("should render a link back to My Workbooks when status is %i", (status) => {
    // given / when
    render(<RouteErrorView error={routeErrorResponse(status)} />);

    // then
    expect(screen.getByTestId("route-error-back")).toHaveAttribute("href", "/workbooks");
  });

  it.each([409, 500, 400])("should render a reload button when status is %i", (status) => {
    // given / when
    render(<RouteErrorView error={routeErrorResponse(status)} />);

    // then
    expect(screen.getByTestId("route-error-reload")).toBeInTheDocument();
  });

  it.each([403, 404])("should not render a reload button when status is %i", (status) => {
    // given / when
    render(<RouteErrorView error={routeErrorResponse(status)} />);

    // then
    expect(screen.queryByTestId("route-error-reload")).not.toBeInTheDocument();
  });

  it("should render the stack of a thrown Error in development", () => {
    // given
    vi.stubEnv("DEV", true);
    const error = new Error("boom");
    error.stack = "Error: boom\n    at stack-marker";

    // when
    render(<RouteErrorView error={error} />);

    // then
    expect(screen.getByTestId("route-error-stack")).toHaveTextContent("stack-marker");
  });

  it("should not render the stack of a thrown Error outside development", () => {
    // given
    vi.stubEnv("DEV", false);
    const error = new Error("boom");

    // when
    render(<RouteErrorView error={error} />);

    // then
    expect(screen.queryByTestId("route-error-stack")).not.toBeInTheDocument();
  });

  it("should render a distinct title for each kind", () => {
    // given
    const statuses = [403, 404, 409, 500, 400];

    // when
    const titles = statuses.map((status) => {
      const { unmount } = render(<RouteErrorView error={routeErrorResponse(status)} />);
      const title = screen.getByRole("heading", { level: 1 }).textContent;
      unmount();
      return title;
    });

    // then
    expect(new Set(titles).size).toBe(statuses.length);
  });
});
