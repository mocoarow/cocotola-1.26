/**
 * Returns "now" rendered as a canonical YYYY-MM-DD string in the supplied
 * IANA timezone. Used as the daily-stats bucket key (X-Local-Date header)
 * sent to the cocotola-question backend on every answer submission and on
 * dashboard reads — the server treats this string as the authoritative
 * user-local "today" so a single user TZ preference governs bucketing
 * regardless of where the request is served from.
 *
 * Uses Intl.DateTimeFormat with `en-CA` because that locale renders the
 * Gregorian calendar in YYYY-MM-DD form without timezone parsing
 * pitfalls (`sv-SE` would also work; we picked one).
 */
export function getLocalDateKey(timezone: string, now: Date = new Date()): string {
  const formatter = new Intl.DateTimeFormat("en-CA", {
    timeZone: timezone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  return formatter.format(now);
}

/**
 * Returns the browser's detected IANA timezone, falling back to the
 * supplied default when the runtime cannot resolve one (server-side
 * rendering or older runtimes).
 */
export function detectBrowserTimezone(fallback = "UTC"): string {
  try {
    const resolved = Intl.DateTimeFormat().resolvedOptions().timeZone;
    return resolved && resolved.length > 0 ? resolved : fallback;
  } catch {
    return fallback;
  }
}

/**
 * Anchored regex enforcing the YYYY-MM-DD shape used everywhere we
 * round-trip a date through HTTP (X-Local-Date header, form fields, etc.).
 * Centralized here so server-side route guards and any future client-side
 * checks cannot drift apart on the wire format.
 */
export const LOCAL_DATE_KEY_PATTERN = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Character-class allow-list for IANA timezone names sent over HTTP
 * (X-Local-Timezone header, form fields, etc.). Mirrors the OpenAPI
 * `pattern` declared on the corresponding backend schemas. Mechanical
 * validity only — the backend additionally resolves the name through
 * Go's `time.LoadLocation` to reject syntactically-valid-but-nonexistent
 * zones (e.g. "Not/AZone"). Defined here so every form action that
 * sanitizes the field reaches for the same regex.
 */
export const TIMEZONE_PATTERN = /^[A-Za-z_/+\-0-9]{1,64}$/;

/** Reports whether v looks like the YYYY-MM-DD pattern accepted on the wire. */
export function isValidLocalDateKey(v: string): boolean {
  return LOCAL_DATE_KEY_PATTERN.test(v);
}

/**
 * Reports whether v passes the character-class shape expected of an IANA
 * timezone name. Real IANA-name resolution happens server-side; this
 * guard exists so client-side route actions do not forward obviously
 * malformed values to the backend.
 */
export function isValidIanaTimezoneShape(v: string): boolean {
  return TIMEZONE_PATTERN.test(v);
}
