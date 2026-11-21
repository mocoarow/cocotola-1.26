import { describe, expect, it } from "vitest";
import { isAccessTokenExpiring } from "./access-token.server";

const NOW_MS = Date.UTC(2026, 10, 21, 12, 0, 0);
const NOW_SEC = NOW_MS / 1000;

function encode(value: string): string {
  return Buffer.from(value).toString("base64url");
}

function jwtWithPayload(payload: string): string {
  return `${encode('{"alg":"HS256","typ":"JWT"}')}.${encode(payload)}.signature`;
}

function jwtExpiringAt(expSec: number): string {
  return jwtWithPayload(JSON.stringify({ exp: expSec }));
}

describe("isAccessTokenExpiring", () => {
  it("should return false when the token expires in more than 60 seconds", () => {
    // given
    const token = jwtExpiringAt(NOW_SEC + 61);

    // when
    const result = isAccessTokenExpiring(token, NOW_MS);

    // then
    expect(result).toBe(false);
  });

  it("should return true when the token expires in exactly 60 seconds", () => {
    // given
    const token = jwtExpiringAt(NOW_SEC + 60);

    // when
    const result = isAccessTokenExpiring(token, NOW_MS);

    // then
    expect(result).toBe(true);
  });

  it("should return true when the token has already expired", () => {
    // given
    const token = jwtExpiringAt(NOW_SEC - 1);

    // when
    const result = isAccessTokenExpiring(token, NOW_MS);

    // then
    expect(result).toBe(true);
  });

  it.each([
    { name: "not three segments", token: "only.two" },
    { name: "payload is not JSON", token: jwtWithPayload("not-json") },
    { name: "payload decodes to empty", token: "header.@@@.signature" },
    { name: "exp is missing", token: jwtWithPayload(JSON.stringify({ sub: "u1" })) },
    {
      name: "exp is a string",
      token: jwtWithPayload(JSON.stringify({ exp: String(NOW_SEC + 3600) })),
    },
    { name: "payload is JSON null", token: jwtWithPayload("null") },
  ])("should return true when the token is unreadable: $name", ({ token }) => {
    // given / when
    const result = isAccessTokenExpiring(token, NOW_MS);

    // then
    expect(result).toBe(true);
  });
});
