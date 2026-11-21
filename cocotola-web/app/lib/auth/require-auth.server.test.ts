import { RouterContextProvider } from "react-router";
import { describe, expect, it } from "vitest";
import { authContext } from "./refresh-middleware.server";
import { requireAuth } from "./require-auth.server";

function thrownBy(fn: () => unknown): unknown {
  try {
    fn();
  } catch (error) {
    return error;
  }
  return undefined;
}

describe("requireAuth", () => {
  it("should throw a redirect to /login when there is no access token", () => {
    // given
    const context = new RouterContextProvider();
    context.set(authContext, null);

    // when
    const thrown = thrownBy(() => requireAuth(context));

    // then
    expect((thrown as Response).headers.get("Location")).toBe("/login");
  });

  it("should return the access token when there is one", () => {
    // given
    const context = new RouterContextProvider();
    context.set(authContext, { accessToken: "at" });

    // when
    const result = requireAuth(context);

    // then
    expect(result).toEqual({ accessToken: "at" });
  });
});
