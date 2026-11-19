import { isRouteErrorResponse } from "react-router";

export type RouteErrorKind = "forbidden" | "notFound" | "conflict" | "server" | "unexpected";

const STATUS_FORBIDDEN = 403;
const STATUS_NOT_FOUND = 404;
const STATUS_CONFLICT = 409;
const STATUS_SERVER_ERROR_MIN = 500;

export function isConflictOrServerError(status: number): boolean {
  return status === STATUS_CONFLICT || status >= STATUS_SERVER_ERROR_MIN;
}

export function classifyRouteError(error: unknown): RouteErrorKind {
  if (!isRouteErrorResponse(error)) return "unexpected";
  if (error.status === STATUS_FORBIDDEN) return "forbidden";
  if (error.status === STATUS_NOT_FOUND) return "notFound";
  if (error.status === STATUS_CONFLICT) return "conflict";
  if (error.status >= STATUS_SERVER_ERROR_MIN) return "server";
  return "unexpected";
}
