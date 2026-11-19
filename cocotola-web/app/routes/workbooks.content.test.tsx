import { render, screen } from "@testing-library/react";
import { createRoutesStub, Outlet } from "react-router";
import { describe, expect, it } from "vitest";

import "~/i18n/config";

import routes from "~/routes";
import WorkbooksContent, { ErrorBoundary } from "./workbooks.content";

const CHILD_PATH = "/workbooks/wb-1/study";

function Layout() {
  return (
    <div>
      <nav data-testid="layout-nav" />
      <Outlet />
    </div>
  );
}

function renderThrowing(thrown: unknown, initialEntry = CHILD_PATH) {
  const Stub = createRoutesStub([
    {
      Component: Layout,
      children: [
        {
          Component: WorkbooksContent,
          ErrorBoundary,
          children: [
            {
              path: "/workbooks/:workbookId/study",
              HydrateFallback: () => null,
              loader: () => {
                throw thrown;
              },
              Component: () => null,
            },
          ],
        },
      ],
    },
  ]);
  render(<Stub initialEntries={[initialEntry]} />);
}

describe("WorkbooksContent ErrorBoundary", () => {
  it.each([
    [403, "forbidden"],
    [404, "notFound"],
    [409, "conflict"],
    [500, "server"],
    [503, "server"],
    [400, "unexpected"],
    [401, "unexpected"],
  ])("should show the %i error as %s", async (status, kind) => {
    // given
    renderThrowing(new Response(null, { status }));

    // when
    const view = await screen.findByTestId(`route-error-${kind}`);

    // then
    expect(view).toBeInTheDocument();
  });

  it("should show a non-response error as unexpected", async () => {
    // given
    renderThrowing(new Error("boom"));

    // when
    const view = await screen.findByTestId("route-error-unexpected");

    // then
    expect(view).toBeInTheDocument();
  });

  it("should keep the layout around the error", async () => {
    // given
    renderThrowing(new Response(null, { status: 500 }));

    // when
    await screen.findByTestId("route-error-server");

    // then
    expect(screen.getByTestId("layout-nav")).toBeInTheDocument();
  });

  it.each([
    409, 500,
  ])("should offer reloading the current URL when the status is %i", async (status) => {
    // given
    renderThrowing(new Response(null, { status }), `${CHILD_PATH}?practice=true`);

    // when
    const reload = await screen.findByTestId("route-error-reload");

    // then
    expect(reload).toHaveAttribute("href", `${CHILD_PATH}?practice=true`);
  });

  it.each([403, 404, 400])("should not offer reloading when the status is %i", async (status) => {
    // given
    renderThrowing(new Response(null, { status }));

    // when
    await screen.findByTestId("route-error-back");

    // then
    expect(screen.queryByTestId("route-error-reload")).not.toBeInTheDocument();
  });

  it.each([
    403, 404, 409, 500,
  ])("should link back to my workbooks when the status is %i", async (status) => {
    // given
    renderThrowing(new Response(null, { status }));

    // when
    const back = await screen.findByTestId("route-error-back");

    // then
    expect(back).toHaveAttribute("href", "/workbooks");
  });
});

describe("routes config", () => {
  it("should nest every workbooks layout child under the content error boundary", () => {
    // given
    const layout = routes.find((r) => r.file === "routes/workbooks.tsx");

    // when
    const children = layout?.children?.map((r) => r.file);

    // then
    expect(children).toEqual(["routes/workbooks.content.tsx"]);
  });
});
