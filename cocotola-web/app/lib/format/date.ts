const LOCALE_MAP: Record<string, string> = {
  ja: "ja-JP",
  en: "en-US",
  ko: "ko-KR",
};

/** Formats an ISO date string into a short, locale-aware "Apr 29, 2026"-style label. */
export function formatDate(dateStr: string, locale: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString(LOCALE_MAP[locale] ?? locale, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
