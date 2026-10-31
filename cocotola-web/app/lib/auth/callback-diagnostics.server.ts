import { parseCookieHeader } from "@supabase/ssr";
import { LOGIN_PROBE_COOKIE_NAME, parseLoginProbe } from "./login-probe";

// Only cookie counts and validated project refs are reported: cookie names are
// client-controlled input and must not reach the logs verbatim.

const SUPABASE_COOKIE_PREFIX = "sb-";
const AUTH_TOKEN_SUFFIX = "-auth-token";
const CODE_VERIFIER_SUFFIX = "-code-verifier";
const VERIFIER_COOKIE_PATTERN = new RegExp(
  `^${SUPABASE_COOKIE_PREFIX}([a-z0-9]{20})${AUTH_TOKEN_SUFFIX}${CODE_VERIFIER_SUFFIX}(?:\\.\\d+)?$`,
);
const MS_PER_SEC = 1000;

type ProbeDiagnostics =
  | { status: "received"; pageAgeSec: number; clickAgeSec: number }
  | { status: "missing" }
  | { status: "malformed" };

type Verdict =
  | "verifier_received"
  | "other_project_verifier"
  | "probe_missing"
  | "verifier_missing_after_click";

export type CallbackCookieDiagnostics = {
  verdict: Verdict;
  cookieCount: number;
  supabaseCookieCount: number;
  hasExpectedVerifier: boolean;
  otherVerifierRefs: string[];
  probe: ProbeDiagnostics;
};

// Mirrors the supabase-js default storageKey; update if cookieOptions.name is set.
export function supabaseAuthStorageKey(supabaseUrl: string): string {
  const ref = new URL(supabaseUrl).hostname.split(".")[0];
  return `${SUPABASE_COOKIE_PREFIX}${ref}${AUTH_TOKEN_SUFFIX}`;
}

export function diagnoseCallbackCookies(
  cookieHeader: string,
  storageKey: string,
  nowMs: number,
): CallbackCookieDiagnostics {
  const cookies = parseCookieHeader(cookieHeader);
  const names = cookies.map((c) => c.name);
  const verifierCookieName = `${storageKey}${CODE_VERIFIER_SUFFIX}`;
  const hasExpectedVerifier = names.some(
    (n) => n === verifierCookieName || n.startsWith(`${verifierCookieName}.`),
  );
  const refs = otherVerifierRefs(names, storageKey);
  const probe = diagnoseProbe(
    cookies.find((c) => c.name === LOGIN_PROBE_COOKIE_NAME)?.value,
    nowMs,
  );

  return {
    verdict: decideVerdict(hasExpectedVerifier, refs, probe),
    cookieCount: names.length,
    supabaseCookieCount: names.filter((n) => n.startsWith(SUPABASE_COOKIE_PREFIX)).length,
    hasExpectedVerifier,
    otherVerifierRefs: refs,
    probe,
  };
}

function decideVerdict(
  hasExpectedVerifier: boolean,
  otherRefs: string[],
  probe: ProbeDiagnostics,
): Verdict {
  if (hasExpectedVerifier) {
    return "verifier_received";
  }
  if (otherRefs.length > 0) {
    return "other_project_verifier";
  }
  if (probe.status === "missing") {
    return "probe_missing";
  }
  return "verifier_missing_after_click";
}

function otherVerifierRefs(names: string[], storageKey: string): string[] {
  const refs = names
    .filter((n) => !n.startsWith(storageKey))
    .map((n) => VERIFIER_COOKIE_PATTERN.exec(n)?.[1])
    .filter((ref): ref is string => ref !== undefined);
  return [...new Set(refs)];
}

function diagnoseProbe(value: string | undefined, nowMs: number): ProbeDiagnostics {
  if (value === undefined) {
    return { status: "missing" };
  }
  const probe = parseLoginProbe(value);
  if (!probe) {
    return { status: "malformed" };
  }
  return {
    status: "received",
    pageAgeSec: Math.floor((nowMs - probe.pageLoadedAtMs) / MS_PER_SEC),
    clickAgeSec: Math.floor((nowMs - probe.clickedAtMs) / MS_PER_SEC),
  };
}
