import { describe, expect, it } from "vitest";
import { classifyRouteError, isConflictOrServerError } from "./route-error";

function routeErrorResponse(status: number) {
  return { status, statusText: "", internal: false, data: "" };
}

describe("classifyRouteError", () => {
  it.each([
    [403, "forbidden"],
    [404, "notFound"],
    [409, "conflict"],
    [500, "server"],
    [503, "server"],
    [400, "unexpected"],
    [401, "unexpected"],
  ] as const)("should classify status %i as %s", (status, kind) => {
    // given
    const error = routeErrorResponse(status);

    // when
    const result = classifyRouteError(error);

    // then
    expect(result).toBe(kind);
  });

  it("should classify a thrown Error as unexpected", () => {
    // given
    const error = new Error("boom");

    // when
    const result = classifyRouteError(error);

    // then
    expect(result).toBe("unexpected");
  });
});

describe("isConflictOrServerError", () => {
  it.each([409, 500, 503])("should return true when status is %i", (status) => {
    // given / when
    const result = isConflictOrServerError(status);

    // then
    expect(result).toBe(true);
  });

  it.each([400, 401, 403, 404])("should return false when status is %i", (status) => {
    // given / when
    const result = isConflictOrServerError(status);

    // then
    expect(result).toBe(false);
  });
});
