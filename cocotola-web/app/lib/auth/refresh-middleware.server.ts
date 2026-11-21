import { createContext, type MiddlewareFunction, type RouterContextProvider } from "react-router";
import { type RefreshResult, refreshAccessToken } from "~/lib/api/auth.server";
import { isAccessTokenExpiring } from "./access-token.server";
import { commitSession, destroySession, getSession, SESSION_COOKIE_NAME } from "./session.server";

export const authContext = createContext<{ accessToken: string } | null>(null);

function hasSessionCookie(response: Response): boolean {
  return response.headers.getSetCookie().some((c) => c.startsWith(`${SESSION_COOKIE_NAME}=`));
}

async function tryRefresh(refreshToken: string): Promise<RefreshResult | undefined> {
  try {
    return await refreshAccessToken(refreshToken);
  } catch (error) {
    console.error("[auth] token refresh failed:", error);
    return undefined;
  }
}

async function applyRefreshResult(
  session: Awaited<ReturnType<typeof getSession>>,
  context: Readonly<RouterContextProvider>,
  result: RefreshResult,
): Promise<string> {
  if (result.kind === "unauthorized") {
    return destroySession(session);
  }
  session.set("accessToken", result.accessToken);
  context.set(authContext, { accessToken: result.accessToken });
  return commitSession(session);
}

export const refreshAccessTokenMiddleware: MiddlewareFunction<Response> = async (
  { request, context },
  next,
) => {
  const session = await getSession(request);
  const accessToken = session.get("accessToken");
  if (!accessToken) {
    return next();
  }

  const refreshToken = session.get("refreshToken");
  const result =
    refreshToken && isAccessTokenExpiring(accessToken, Date.now())
      ? await tryRefresh(refreshToken)
      : undefined;
  if (result === undefined) {
    context.set(authContext, { accessToken });
    return next();
  }

  const setCookie = await applyRefreshResult(session, context, result);
  const response = await next();
  if (!hasSessionCookie(response)) {
    response.headers.append("Set-Cookie", setCookie);
  }
  return response;
};
