import { isRouteErrorResponse } from "react-router";
import { isServerErrorStatus } from "./http-status";

export type RouteErrorKind = "forbidden" | "notFound" | "conflict" | "server" | "unexpected";

const STATUS_FORBIDDEN = 403;
const STATUS_NOT_FOUND = 404;
const STATUS_CONFLICT = 409;
const STATUS_REDIRECT_MIN = 300;
const STATUS_REDIRECT_MAX_EXCLUSIVE = 400;

export function isRedirectResponse(error: unknown): error is Response {
  return (
    error instanceof Response &&
    error.status >= STATUS_REDIRECT_MIN &&
    error.status < STATUS_REDIRECT_MAX_EXCLUSIVE
  );
}

export function isConflictOrServerError(status: number): boolean {
  return status === STATUS_CONFLICT || isServerErrorStatus(status);
}

export function classifyRouteError(error: unknown): RouteErrorKind {
  if (!isRouteErrorResponse(error)) return "unexpected";
  if (error.status === STATUS_FORBIDDEN) return "forbidden";
  if (error.status === STATUS_NOT_FOUND) return "notFound";
  if (error.status === STATUS_CONFLICT) return "conflict";
  if (isServerErrorStatus(error.status)) return "server";
  return "unexpected";
}
