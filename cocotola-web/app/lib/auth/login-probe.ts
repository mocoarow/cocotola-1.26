// Marker cookie written when the Google login button is clicked so that
// /auth/callback can tell whether cookies from the click reached it.

export const LOGIN_PROBE_COOKIE_NAME = "cocotola_login_probe";

const LOGIN_PROBE_MAX_AGE_MS = 10 * 60 * 1000;
const LOGIN_PROBE_VALUE_PATTERN = /^(\d+)\.(\d+)$/;

export type LoginProbe = {
  pageLoadedAtMs: number;
  clickedAtMs: number;
};

export function buildLoginProbeCookie(probe: LoginProbe): CookieInit {
  return {
    name: LOGIN_PROBE_COOKIE_NAME,
    value: `${probe.pageLoadedAtMs}.${probe.clickedAtMs}`,
    path: "/",
    sameSite: "lax",
    expires: probe.clickedAtMs + LOGIN_PROBE_MAX_AGE_MS,
  };
}

// Best effort: login must proceed even when the probe cannot be written.
export async function writeLoginProbe(probe: LoginProbe): Promise<void> {
  if (typeof cookieStore === "undefined") {
    return;
  }
  try {
    await cookieStore.set(buildLoginProbeCookie(probe));
  } catch (e) {
    console.warn("login: write login probe failed", e);
  }
}

export function parseLoginProbe(value: string): LoginProbe | null {
  const match = LOGIN_PROBE_VALUE_PATTERN.exec(value);
  if (!match) {
    return null;
  }
  return { pageLoadedAtMs: Number(match[1]), clickedAtMs: Number(match[2]) };
}
