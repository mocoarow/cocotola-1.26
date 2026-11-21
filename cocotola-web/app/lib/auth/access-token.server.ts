const REFRESH_SKEW_MS = 60_000;

function readExpSec(accessToken: string): number | undefined {
  const segments = accessToken.split(".");
  if (segments.length !== 3) return undefined;

  try {
    const payload: unknown = JSON.parse(Buffer.from(segments[1], "base64url").toString("utf8"));
    if (typeof payload !== "object" || payload === null || !("exp" in payload)) return undefined;
    return typeof payload.exp === "number" ? payload.exp : undefined;
  } catch {
    return undefined;
  }
}

/** The signature is not verified; the backend does that. Unreadable tokens count as expiring. */
export function isAccessTokenExpiring(accessToken: string, nowMs: number): boolean {
  const expSec = readExpSec(accessToken);
  if (expSec === undefined) return true;
  return expSec * 1000 - nowMs <= REFRESH_SKEW_MS;
}
