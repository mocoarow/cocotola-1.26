import { describe, expect, it } from "vitest";
import { diagnoseCallbackCookies, supabaseAuthStorageKey } from "./callback-diagnostics.server";

const expectedRef = "abcdefghijklmnopqrst";
const otherRef = "zyxwvutsrqponmlkjihg";
const storageKey = `sb-${expectedRef}-auth-token`;
const verifierCookieName = `${storageKey}-code-verifier`;
const flowId = "0123456789abcdef0123456789abcdef";
const nowMs = 100_000;

describe("supabaseAuthStorageKey", () => {
  it("should derive the storage key from the first hostname label", () => {
    // given
    const supabaseUrl = `https://${expectedRef}.supabase.co`;

    // when
    const key = supabaseAuthStorageKey(supabaseUrl);

    // then
    expect(key).toBe(storageKey);
  });
});

describe("diagnoseCallbackCookies", () => {
  it("should report the expected verifier when it is received", () => {
    // given
    const cookieHeader = `${verifierCookieName}=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.hasExpectedVerifier).toBe(true);
  });

  it("should report the expected verifier when it is received as a chunk", () => {
    // given
    const cookieHeader = `${verifierCookieName}.0=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.hasExpectedVerifier).toBe(true);
  });

  it("should report no expected verifier when it is missing", () => {
    // given
    const cookieHeader = "other=v";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.hasExpectedVerifier).toBe(false);
  });

  it("should not treat a flow slot as the expected verifier", () => {
    // given
    const cookieHeader = `${storageKey}-flow-${flowId}-code-verifier=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.hasExpectedVerifier).toBe(false);
  });

  it.each([
    {
      verdict: "verifier_received",
      cookieHeader: `${verifierCookieName}=v; sb-${otherRef}-auth-token-code-verifier=v`,
    },
    {
      verdict: "other_project_verifier",
      cookieHeader: `sb-${otherRef}-auth-token-code-verifier=v; cocotola_login_probe=1.2`,
    },
    {
      verdict: "probe_missing",
      cookieHeader: `${storageKey}-flow-${flowId}-code-verifier=v`,
    },
    {
      verdict: "verifier_missing_after_click",
      cookieHeader: "cocotola_login_probe=1.2",
    },
  ])("should decide $verdict from the received cookies", ({ verdict, cookieHeader }) => {
    // given (cookieHeader)

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.verdict).toBe(verdict);
  });

  it("should list refs of verifiers for other projects", () => {
    // given
    const cookieHeader = `${verifierCookieName}=v; sb-${otherRef}-auth-token-code-verifier=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.otherVerifierRefs).toEqual([otherRef]);
  });

  it("should list a ref once when its verifier is chunked", () => {
    // given
    const cookieHeader = `sb-${otherRef}-auth-token-code-verifier.0=v; sb-${otherRef}-auth-token-code-verifier.1=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.otherVerifierRefs).toEqual([otherRef]);
  });

  it("should not list the expected project when its verifier is chunked", () => {
    // given
    const cookieHeader = `${verifierCookieName}.0=v`;

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.otherVerifierRefs).toEqual([]);
  });

  it("should ignore verifier-like names whose ref is not a project ref", () => {
    // given
    const cookieHeader =
      "sb-<script>-auth-token-code-verifier=v; sb-short-auth-token-code-verifier=v";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.otherVerifierRefs).toEqual([]);
  });

  it("should count all cookies", () => {
    // given
    const cookieHeader = "a=1; sb-x-auth-token=2; b=3";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.cookieCount).toBe(3);
  });

  it("should count cookies prefixed with sb-", () => {
    // given
    const cookieHeader = "a=1; sb-x-auth-token=2; sb-y=3";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.supabaseCookieCount).toBe(2);
  });

  it("should report ages in seconds when the probe is received", () => {
    // given
    const cookieHeader = "cocotola_login_probe=40000.97500";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.probe).toEqual({ status: "received", pageAgeSec: 60, clickAgeSec: 2 });
  });

  it("should report a missing probe when it is not received", () => {
    // given
    const cookieHeader = "a=1";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.probe).toEqual({ status: "missing" });
  });

  it("should report a malformed probe when its value cannot be parsed", () => {
    // given
    const cookieHeader = "cocotola_login_probe=broken";

    // when
    const result = diagnoseCallbackCookies(cookieHeader, storageKey, nowMs);

    // then
    expect(result.probe).toEqual({ status: "malformed" });
  });
});
