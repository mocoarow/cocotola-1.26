export const CONTENT_LANGUAGES = ["ja", "en", "it", "fr", "de", "es", "zh", "ko", "pt"] as const;

export type ContentLanguage = (typeof CONTENT_LANGUAGES)[number];

export const DEFAULT_CONTENT_LANGUAGE: ContentLanguage = "ja";

export function isContentLanguage(value: string): value is ContentLanguage {
  return (CONTENT_LANGUAGES as readonly string[]).includes(value);
}

export function toContentLanguage(value: string | undefined): ContentLanguage {
  return value !== undefined && isContentLanguage(value) ? value : DEFAULT_CONTENT_LANGUAGE;
}
