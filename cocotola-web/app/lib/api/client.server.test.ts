import { format } from "node:util";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { catchError, jsonResponse, newClient } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";
const AUTH_BASE_URL = "http://localhost:8080";

function lastFetchInit(): RequestInit {
  return fetchMock.mock.lastCall?.[1] as RequestInit;
}

const baseOptions = { source: "test:call", errorMessage: "Failed to call" };
const identity = (raw: unknown) => raw;

describe("createBackendClient", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", QUESTION_BASE_URL);
    vi.stubEnv("AUTH_BASE_URL", AUTH_BASE_URL);
    vi.spyOn(console, "info").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
    fetchMock.mockReset();
  });

  it("should build URL from QUESTION_BASE_URL and path when service is question", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().get("question", "/api/v1/workbook?spaceId=sp-1", baseOptions);

    // then
    expect(fetchMock.mock.calls[0][0]).toBe(`${QUESTION_BASE_URL}/api/v1/workbook?spaceId=sp-1`);
  });

  it("should build URL from AUTH_BASE_URL and path when service is auth", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().get("auth", "/api/v1/auth/me", baseOptions);

    // then
    expect(fetchMock.mock.calls[0][0]).toBe(`${AUTH_BASE_URL}/api/v1/auth/me`);
  });

  it.each([
    ["get", "GET"],
    ["post", "POST"],
    ["put", "PUT"],
    ["delete", "DELETE"],
  ] as const)("should send %s as HTTP method %s", async (verb, method) => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient()[verb]("question", "/api/v1/x", baseOptions);

    // then
    expect(lastFetchInit().method).toBe(method);
  });

  it("should send only the Bearer access token header when no body or headers are given", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().get("question", "/api/v1/x", baseOptions);

    // then
    expect(lastFetchInit().headers).toEqual({ Authorization: "Bearer test-token" });
  });

  it("should send Content-Type application/json when body is given", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().post("question", "/api/v1/x", { ...baseOptions, body: { title: "t" } });

    // then
    expect(lastFetchInit().headers).toMatchObject({ "Content-Type": "application/json" });
  });

  it("should send the body serialized as JSON when body is given", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().post("question", "/api/v1/x", { ...baseOptions, body: { title: "t" } });

    // then
    expect(lastFetchInit().body).toBe(JSON.stringify({ title: "t" }));
  });

  it("should send extra headers when headers are given", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().post("question", "/api/v1/x", {
      ...baseOptions,
      headers: { "X-Local-Timezone": "America/Los_Angeles" },
    });

    // then
    expect(lastFetchInit().headers).toMatchObject({ "X-Local-Timezone": "America/Los_Angeles" });
  });

  it("should send the access token when headers also carry Authorization", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await newClient().get("question", "/api/v1/x", {
      ...baseOptions,
      headers: { Authorization: "Bearer other-token" },
    });

    // then
    expect(lastFetchInit().headers).toMatchObject({ Authorization: "Bearer test-token" });
  });

  it("should throw redirect to /login when backend returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 401 }));

    // when
    const caught = await catchError(newClient().get("question", "/api/v1/x", baseOptions));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should destroy the session cookie when backend returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 401 }));

    // when
    const caught = await catchError(newClient().get("question", "/api/v1/x", baseOptions));

    // then
    expect((caught as Response).headers.get("Set-Cookie")).toMatch(/__cocotola_session=;/);
  });

  it("should throw Response with the backend status when backend returns non-2xx", async () => {
    // given
    fetchMock.mockResolvedValue(new Response("boom", { status: 503 }));

    // when
    const caught = await catchError(newClient().get("question", "/api/v1/x", baseOptions));

    // then
    expect((caught as Response).status).toBe(503);
  });

  it("should throw Response with errorMessage as body when backend returns non-2xx", async () => {
    // given
    fetchMock.mockResolvedValue(new Response("internal detail", { status: 503 }));

    // when
    const caught = await catchError(newClient().get("question", "/api/v1/x", baseOptions));

    // then
    expect(await (caught as Response).text()).toBe("Failed to call");
  });

  it("should log source, status and url when backend returns non-2xx", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 503 }));

    // when
    await catchError(newClient().get("question", "/api/v1/x", baseOptions));

    // then
    const logged = vi.mocked(console.error).mock.calls.map((args) => format(...args));
    expect(logged).toContain(`[test:call] failed: status=503, url=${QUESTION_BASE_URL}/api/v1/x`);
  });

  it("should throw the error returned by mapError when backend returns non-2xx", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));
    const mapped = new Error("mapped");

    // when
    const caught = await catchError(
      newClient().post("question", "/api/v1/x", { ...baseOptions, mapError: () => mapped }),
    );

    // then
    expect(caught).toBe(mapped);
  });

  it("should pass status and errorMessage to mapError when backend returns non-2xx", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 400 }));
    const mapError = vi.fn(() => new Error("mapped"));

    // when
    await catchError(newClient().post("question", "/api/v1/x", { ...baseOptions, mapError }));

    // then
    expect(mapError).toHaveBeenCalledWith(400, "Failed to call");
  });

  it("should throw the default Response when mapError returns undefined", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when
    const caught = await catchError(
      newClient().post("question", "/api/v1/x", { ...baseOptions, mapError: () => undefined }),
    );

    // then
    expect((caught as Response).status).toBe(500);
  });

  it("should return the value produced by parse when backend returns 2xx", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ workbooks: [{ workbookId: "wb-1" }] }));

    // when
    const result = await newClient().get("question", "/api/v1/x", {
      ...baseOptions,
      parse: (raw) => (raw as { workbooks: unknown[] }).workbooks.length,
    });

    // then
    expect(result).toBe(1);
  });

  it("should throw 502 Response when body is not JSON", async () => {
    // given
    fetchMock.mockResolvedValue(new Response("<html>oops</html>", { status: 200 }));

    // when
    const caught = await catchError(
      newClient().get("question", "/api/v1/x", { ...baseOptions, parse: identity }),
    );

    // then
    expect((caught as Response).status).toBe(502);
  });

  it("should throw 502 Response when parse throws", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ unexpected: true }));

    // when
    const caught = await catchError(
      newClient().get("question", "/api/v1/x", {
        ...baseOptions,
        parse: () => {
          throw new Error("invalid shape");
        },
      }),
    );

    // then
    expect((caught as Response).status).toBe(502);
  });

  it("should log source, error class and url when parse throws", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ unexpected: true }));

    // when
    await catchError(
      newClient().get("question", "/api/v1/x", {
        ...baseOptions,
        parse: () => {
          throw new TypeError("invalid shape");
        },
      }),
    );

    // then
    const logged = vi.mocked(console.error).mock.calls.map((args) => format(...args));
    expect(logged).toContain(
      `[test:call] invalid response: error=TypeError, url=${QUESTION_BASE_URL}/api/v1/x`,
    );
  });

  it("should not log the parser error message when parse throws", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ loginId: "user@example.com" }));

    // when
    await catchError(
      newClient().get("question", "/api/v1/x", {
        ...baseOptions,
        parse: () => {
          throw new Error("unexpected loginId user@example.com");
        },
      }),
    );

    // then
    const logged = vi.mocked(console.error).mock.calls.map((args) => format(...args));
    expect(logged.join("\n")).not.toContain("user@example.com");
  });

  it("should resolve with undefined when parse is omitted and backend returns no body", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    const result = await newClient().delete("question", "/api/v1/x", baseOptions);

    // then
    expect(result).toBeUndefined();
  });
});
