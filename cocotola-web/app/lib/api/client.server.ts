import { redirectOnUnauthorized } from "~/lib/auth/session.server";

import { fetchWithIdToken, getAuthUrl, getQuestionUrl } from "./fetch.server";

const BAD_GATEWAY = 502;

export type BackendService = "question" | "auth";

export type ResponseParser<T> = (raw: unknown) => T;

/** Returning undefined falls back to `new Response(errorMessage, { status })`. */
export type ErrorMapper = (status: number, errorMessage: string) => Error | Response | undefined;

export type CallOptions = {
  /** Log tag such as "workbook:listWorkbooks". */
  source: string;
  errorMessage: string;
  headers?: Record<string, string>;
  mapError?: ErrorMapper;
};

export type BodyCallOptions = CallOptions & { body?: unknown };

export type BackendCall<O extends CallOptions> = {
  <T>(service: BackendService, path: string, options: O & { parse: ResponseParser<T> }): Promise<T>;
  (service: BackendService, path: string, options: O): Promise<void>;
};

export type BackendClient = {
  get: BackendCall<CallOptions>;
  post: BackendCall<BodyCallOptions>;
  put: BackendCall<BodyCallOptions>;
  delete: BackendCall<CallOptions>;
};

type HttpMethod = "GET" | "POST" | "PUT" | "DELETE";

type ClientContext = {
  request: Request;
  accessToken: string;
};

type SendOptions = BodyCallOptions & { parse?: ResponseParser<unknown> };

/**
 * Takes the token explicitly instead of calling requireAuth(context):
 * auth.callback.tsx calls backends with a freshly exchanged token that is
 * not in the context yet.
 */
export function createBackendClient(ctx: ClientContext): BackendClient {
  return {
    get: bind<CallOptions>(ctx, "GET"),
    post: bind<BodyCallOptions>(ctx, "POST"),
    put: bind<BodyCallOptions>(ctx, "PUT"),
    delete: bind<CallOptions>(ctx, "DELETE"),
  };
}

function bind<O extends CallOptions>(ctx: ClientContext, method: HttpMethod): BackendCall<O> {
  function call<T>(
    service: BackendService,
    path: string,
    options: O & { parse: ResponseParser<T> },
  ): Promise<T>;
  function call(service: BackendService, path: string, options: O): Promise<void>;
  function call(service: BackendService, path: string, options: SendOptions): Promise<unknown> {
    return send(ctx, method, service, path, options);
  }
  return call;
}

function resolveBaseUrl(service: BackendService): string {
  return service === "question" ? getQuestionUrl() : getAuthUrl();
}

function buildInit(ctx: ClientContext, method: HttpMethod, options: SendOptions): RequestInit {
  const hasBody = options.body !== undefined;
  const headers: Record<string, string> = {
    ...options.headers,
    Authorization: `Bearer ${ctx.accessToken}`,
    ...(hasBody ? { "Content-Type": "application/json" } : {}),
  };
  return hasBody ? { method, headers, body: JSON.stringify(options.body) } : { method, headers };
}

async function send(
  ctx: ClientContext,
  method: HttpMethod,
  service: BackendService,
  path: string,
  options: SendOptions,
): Promise<unknown> {
  const baseUrl = resolveBaseUrl(service);
  const url = `${baseUrl}${path}`;

  const response = await redirectOnUnauthorized(
    ctx.request,
    await fetchWithIdToken(baseUrl, url, buildInit(ctx, method, options)),
    options.source,
  );

  if (!response.ok) {
    console.error(`[${options.source}] failed: status=${response.status}, url=${url}`);
    throw (
      options.mapError?.(response.status, options.errorMessage) ??
      new Response(options.errorMessage, { status: response.status })
    );
  }

  if (!options.parse) return undefined;
  return parseBody(response, url, options.parse, options);
}

async function parseBody<T>(
  response: Response,
  url: string,
  parse: ResponseParser<T>,
  options: CallOptions,
): Promise<T> {
  try {
    return parse(await response.json());
  } catch (error) {
    // Only the error class is logged: parser messages can echo response values.
    const errorName = error instanceof Error ? error.name : typeof error;
    console.error(`[${options.source}] invalid response: error=${errorName}, url=${url}`);
    throw new Response(options.errorMessage, { status: BAD_GATEWAY });
  }
}
