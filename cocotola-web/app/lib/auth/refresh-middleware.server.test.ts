import { RouterContextProvider } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("~/lib/api/auth.server", () => ({
  exchangeSupabaseToken: vi.fn(),
  refreshAccessToken: vi.fn(),
}));

vi.mock("./access-token.server", () => ({
  isAccessTokenExpiring: vi.fn(),
}));

import { refreshAccessToken } from "~/lib/api/auth.server";
import { isAccessTokenExpiring } from "./access-token.server";
import { authContext, refreshAccessTokenMiddleware } from "./refresh-middleware.server";
import { commitSession, getSession, SESSION_COOKIE_NAME } from "./session.server";

const mockedRefresh = vi.mocked(refreshAccessToken);
const mockedIsExpiring = vi.mocked(isAccessTokenExpiring);

async function requestWithSession(tokens: {
  accessToken?: string;
  refreshToken?: string;
}): Promise<Request> {
  const session = await getSession(new Request("http://localhost/"));
  if (tokens.accessToken) session.set("accessToken", tokens.accessToken);
  if (tokens.refreshToken) session.set("refreshToken", tokens.refreshToken);
  const cookie = (await commitSession(session)).split(";")[0];
  return new Request("http://localhost/workbooks", { headers: { Cookie: cookie } });
}

async function runMiddleware(
  request: Request,
  response: Response = new Response(null),
): Promise<{ context: RouterContextProvider; next: ReturnType<typeof vi.fn>; result: Response }> {
  const context = new RouterContextProvider();
  const next = vi.fn(async () => response);
  const result = (await refreshAccessTokenMiddleware(
    { request, url: new URL(request.url), pattern: "/workbooks", params: {}, context },
    next,
  )) as Response;
  return { context, next, result };
}

function sessionCookies(response: Response): string[] {
  return response.headers.getSetCookie().filter((c) => c.startsWith(`${SESSION_COOKIE_NAME}=`));
}

async function sessionValueInCookie(
  setCookie: string,
  key: "accessToken" | "refreshToken",
): Promise<string | undefined> {
  const request = new Request("http://localhost/", {
    headers: { Cookie: setCookie.split(";")[0] },
  });
  return (await getSession(request)).get(key);
}

function responseWithCookies(...cookies: string[]): Response {
  const headers = new Headers();
  for (const cookie of cookies) headers.append("Set-Cookie", cookie);
  return new Response(null, { headers });
}

describe("refreshAccessTokenMiddleware", () => {
  afterEach(() => {
    vi.resetAllMocks();
  });

  it("should not refresh when the access token is not expiring", async () => {
    // given
    mockedIsExpiring.mockReturnValue(false);
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    await runMiddleware(request);

    // then
    expect(mockedRefresh).not.toHaveBeenCalled();
  });

  it("should expose the current access token when it is not expiring", async () => {
    // given
    mockedIsExpiring.mockReturnValue(false);
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toEqual({ accessToken: "old-at" });
  });

  it("should check the expiry of the session access token", async () => {
    // given
    mockedIsExpiring.mockReturnValue(false);
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    await runMiddleware(request);

    // then
    expect(mockedIsExpiring).toHaveBeenCalledWith("old-at", expect.any(Number));
  });

  it("should not set the session cookie when the access token is not expiring", async () => {
    // given
    mockedIsExpiring.mockReturnValue(false);
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request);

    // then
    expect(sessionCookies(result)).toEqual([]);
  });

  it("should refresh with the session refresh token when the access token is expiring", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    await runMiddleware(request);

    // then
    expect(mockedRefresh).toHaveBeenCalledWith("rt");
  });

  it("should expose the refreshed access token when the refresh succeeds", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toEqual({ accessToken: "new-at" });
  });

  it("should store the refreshed access token in the session cookie when the refresh succeeds", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request);

    // then
    const cookies = sessionCookies(result);
    expect(cookies).toHaveLength(1);
    expect(await sessionValueInCookie(cookies[0], "accessToken")).toBe("new-at");
  });

  it("should keep the refresh token in the session cookie when the refresh succeeds", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request);

    // then
    expect(await sessionValueInCookie(sessionCookies(result)[0], "refreshToken")).toBe("rt");
  });

  it("should not override a session cookie set by the route when the refresh succeeds", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });
    const routeCookie = `${SESSION_COOKIE_NAME}=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT`;

    // when
    const { result } = await runMiddleware(request, responseWithCookies(routeCookie));

    // then
    expect(sessionCookies(result)).toEqual([routeCookie]);
  });

  it("should add the session cookie next to cookies with other names", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "refreshed", accessToken: "new-at" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request, responseWithCookies("sb-auth=x; Path=/"));

    // then
    expect(sessionCookies(result)).toHaveLength(1);
  });

  it("should continue to the route when the refresh is unauthorized", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "unauthorized" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { next } = await runMiddleware(request);

    // then
    expect(next).toHaveBeenCalledOnce();
  });

  it("should expose no access token when the refresh is unauthorized", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "unauthorized" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toBeNull();
  });

  it("should destroy the session cookie when the refresh is unauthorized", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "unauthorized" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request);

    // then
    expect(sessionCookies(result)).toEqual([
      expect.stringContaining("Expires=Thu, 01 Jan 1970 00:00:00 GMT"),
    ]);
  });

  it("should not override a session cookie set by the route when the refresh is unauthorized", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockResolvedValue({ kind: "unauthorized" });
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });
    const routeCookie = `${SESSION_COOKIE_NAME}=fresh-session; Path=/`;

    // when
    const { result } = await runMiddleware(request, responseWithCookies(routeCookie));

    // then
    expect(sessionCookies(result)).toEqual([routeCookie]);
  });

  it("should expose the current access token when the refresh throws", async () => {
    // given
    vi.spyOn(console, "error").mockImplementation(() => {});
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockRejectedValue(new Error("Token refresh failed (500)"));
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toEqual({ accessToken: "old-at" });
  });

  it("should not set the session cookie when the refresh throws", async () => {
    // given
    vi.spyOn(console, "error").mockImplementation(() => {});
    mockedIsExpiring.mockReturnValue(true);
    mockedRefresh.mockRejectedValue(new Error("Token refresh failed (500)"));
    const request = await requestWithSession({ accessToken: "old-at", refreshToken: "rt" });

    // when
    const { result } = await runMiddleware(request);

    // then
    expect(sessionCookies(result)).toEqual([]);
  });

  it("should not refresh when the session has no access token", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    const request = await requestWithSession({ refreshToken: "rt" });

    // when
    await runMiddleware(request);

    // then
    expect(mockedRefresh).not.toHaveBeenCalled();
  });

  it("should expose no access token when the session has no access token", async () => {
    // given
    const request = await requestWithSession({ refreshToken: "rt" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toBeNull();
  });

  it("should expose the current access token when the session has no refresh token", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    const request = await requestWithSession({ accessToken: "old-at" });

    // when
    const { context } = await runMiddleware(request);

    // then
    expect(context.get(authContext)).toEqual({ accessToken: "old-at" });
  });

  it("should not refresh when the session has no refresh token", async () => {
    // given
    mockedIsExpiring.mockReturnValue(true);
    const request = await requestWithSession({ accessToken: "old-at" });

    // when
    await runMiddleware(request);

    // then
    expect(mockedRefresh).not.toHaveBeenCalled();
  });
});
