import { expect, type APIRequestContext } from "@playwright/test";

type PasswordCredentials = {
  loginId: string;
  password: string;
  organizationName: string;
};

type GuestCredentials = {
  organizationName: string;
};

type CreateUserInput = {
  ownerToken: string;
  loginId: string;
  password: string;
};

type JwtPayload = {
  exp?: number;
};

function decodeBase64Url(segment: string): string {
  const padded = segment.padEnd(segment.length + ((4 - (segment.length % 4)) % 4), "=");
  const base64 = padded.replace(/-/g, "+").replace(/_/g, "/");
  return Buffer.from(base64, "base64").toString("utf8");
}

function assertValidAccessToken(token: string): void {
  const segments = token.split(".");
  expect(segments).toHaveLength(3);
  const payloadSegment = segments[1];
  expect(payloadSegment).toBeTruthy();
  if (!payloadSegment) return;
  const payload = JSON.parse(decodeBase64Url(payloadSegment)) as JwtPayload;
  expect(payload.exp).toBeDefined();
  if (payload.exp === undefined) return;
  const expMs = payload.exp * 1000;
  expect(expMs).toBeGreaterThan(Date.now());
}

export async function authenticatePassword(
  request: APIRequestContext,
  credentials: PasswordCredentials,
): Promise<string> {
  const response = await request.post("/api/v1/auth/authenticate", {
    data: credentials,
  });
  expect(response.status()).toBe(200);
  const body = (await response.json()) as { accessToken: string };
  assertValidAccessToken(body.accessToken);
  return body.accessToken;
}

export async function authenticateGuest(
  request: APIRequestContext,
  credentials: GuestCredentials,
): Promise<string> {
  const response = await request.post("/api/v1/auth/guest/authenticate", {
    data: credentials,
  });
  expect(response.status()).toBe(200);
  const body = (await response.json()) as { accessToken: string };
  assertValidAccessToken(body.accessToken);
  return body.accessToken;
}

export async function createUser(
  request: APIRequestContext,
  input: CreateUserInput,
): Promise<string> {
  const response = await request.post("/api/v1/auth/user", {
    headers: {
      authorization: `Bearer ${input.ownerToken}`,
    },
    data: {
      loginId: input.loginId,
      password: input.password,
    },
  });
  expect(response.status()).toBe(201);
  const body = (await response.json()) as { appUserId: string };
  return body.appUserId;
}

export function bearer(token: string): Record<string, string> {
  return { authorization: `Bearer ${token}` };
}

export async function updateUserLanguage(
  request: APIRequestContext,
  userToken: string,
  language: "en" | "ja" | "ko",
): Promise<void> {
  const response = await request.put("/api/v1/auth/user-setting/language", {
    headers: bearer(userToken),
    data: { language },
  });
  expect(response.status()).toBe(204);
}
