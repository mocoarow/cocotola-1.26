import { RouterContextProvider } from "react-router";
import { describe, expect, it } from "vitest";
import { authContext } from "~/lib/auth/refresh-middleware.server";
import { loader } from "./login";

type LoaderArgs = Parameters<typeof loader>[0];

function loaderArgs(context: RouterContextProvider): LoaderArgs {
  const request = new Request("http://localhost/login");
  return { request, params: {}, context } as unknown as LoaderArgs;
}

describe("login loader", () => {
  it("should redirect to / when the middleware exposes an access token", async () => {
    // given
    const context = new RouterContextProvider();
    context.set(authContext, { accessToken: "at" });

    // when
    const result = await loader(loaderArgs(context));

    // then
    expect((result as Response).headers.get("Location")).toBe("/");
  });

  it("should render the login page when the middleware exposes no access token", async () => {
    // given
    const context = new RouterContextProvider();
    context.set(authContext, null);

    // when
    const result = await loader(loaderArgs(context));

    // then
    expect(result).toBeNull();
  });
});
