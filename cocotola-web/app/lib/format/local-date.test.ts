import { afterEach, describe, expect, it, vi } from "vitest";

import { detectBrowserTimezone, getLocalDateKey } from "./local-date";

describe("getLocalDateKey", () => {
  it("should_returnTodayInJST_when_callerIsInTokyo", () => {
    // given
    const utcMoment = new Date("2026-06-14T23:30:00Z"); // 08:30 next day in Tokyo

    // when
    const got = getLocalDateKey("Asia/Tokyo", utcMoment);

    // then
    expect(got).toBe("2026-06-15");
  });

  it("should_returnSameYearMonthDay_when_localBoundaryDiffersFromUTC", () => {
    // given
    const utcMoment = new Date("2026-06-15T05:30:00Z"); // 22:30 on the 14th in LA

    // when
    const got = getLocalDateKey("America/Los_Angeles", utcMoment);

    // then
    expect(got).toBe("2026-06-14");
  });

  it("should_returnUTCDay_when_timezoneIsUTC", () => {
    // given
    const utcMoment = new Date("2026-06-14T12:00:00Z");

    // when
    const got = getLocalDateKey("UTC", utcMoment);

    // then
    expect(got).toBe("2026-06-14");
  });

  it("should_returnCanonicalFormat_when_called", () => {
    // given
    const utcMoment = new Date("2026-01-05T00:00:00Z");

    // when
    const got = getLocalDateKey("UTC", utcMoment);

    // then
    expect(got).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(got).toBe("2026-01-05");
  });
});

describe("detectBrowserTimezone", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should_returnIANAName_when_environmentSupportsIntl", () => {
    // when
    const got = detectBrowserTimezone("UTC");

    // then
    expect(typeof got).toBe("string");
    expect(got.length).toBeGreaterThan(0);
  });

  it("should_returnFallback_when_resolvedOptionsTimeZoneIsEmpty", () => {
    // given
    // Some older runtimes (or strict CSP / isolated sandboxes) return an
    // empty timezone from resolvedOptions() instead of throwing — assert
    // that the empty-string branch falls through to the fallback so the
    // dashboard does not silently treat "" as a valid IANA name.
    vi.spyOn(Intl, "DateTimeFormat").mockImplementation(
      () =>
        ({
          resolvedOptions: () =>
            ({
              timeZone: "",
            }) as ReturnType<Intl.DateTimeFormat["resolvedOptions"]>,
        }) as unknown as Intl.DateTimeFormat,
    );

    // when
    const got = detectBrowserTimezone("Asia/Tokyo");

    // then
    expect(got).toBe("Asia/Tokyo");
  });

  it("should_returnFallback_when_intlConstructorThrows", () => {
    // given
    // SSR-style environments without Intl support (or surfaces where the
    // function runs before polyfills load) cause the constructor itself
    // to throw. The function is documented to swallow that and return
    // the caller-supplied fallback rather than crashing the loader.
    vi.spyOn(Intl, "DateTimeFormat").mockImplementation(() => {
      throw new Error("Intl unavailable");
    });

    // when
    const got = detectBrowserTimezone("Asia/Tokyo");

    // then
    expect(got).toBe("Asia/Tokyo");
  });

  it("should_returnDefaultUTC_when_fallbackIsOmittedAndIntlThrows", () => {
    // given
    vi.spyOn(Intl, "DateTimeFormat").mockImplementation(() => {
      throw new Error("Intl unavailable");
    });

    // when
    const got = detectBrowserTimezone();

    // then
    expect(got).toBe("UTC");
  });
});
