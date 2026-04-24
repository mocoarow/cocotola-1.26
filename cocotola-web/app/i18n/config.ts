import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
import en from "./locales/en.json";
import ja from "./locales/ja.json";
import ko from "./locales/ko.json";

export const supportedLanguages = ["en", "ja", "ko"] as const;
export type SupportedLanguage = (typeof supportedLanguages)[number];

const isServer = typeof window === "undefined";

const instance = i18n.use(initReactI18next);

if (!isServer) {
  instance.use(LanguageDetector);
}

instance.init({
  resources: {
    en: { translation: en },
    ja: { translation: ja },
    ko: { translation: ko },
  },
  fallbackLng: "en",
  detection: {
    order: ["cookie", "localStorage", "navigator"],
    caches: ["localStorage", "cookie"],
    lookupLocalStorage: "i18nextLng",
    lookupCookie: "i18nextLng",
    cookieMinutes: 525600,
  },
  interpolation: {
    escapeValue: false,
  },
});

export function createServerInstance(lng: SupportedLanguage) {
  return i18n.cloneInstance({ lng });
}

export function detectLanguageFromRequest(request: Request): SupportedLanguage {
  const cookie = request.headers.get("Cookie") ?? "";
  const match = cookie.match(/i18nextLng=([^;]+)/);
  if (match) {
    const lang = match[1];
    if (supportedLanguages.includes(lang as SupportedLanguage)) {
      return lang as SupportedLanguage;
    }
  }

  const acceptLanguage = request.headers.get("Accept-Language") ?? "";
  for (const part of acceptLanguage.split(",")) {
    const lang = part.trim().split(";")[0].split("-")[0];
    if (supportedLanguages.includes(lang as SupportedLanguage)) {
      return lang as SupportedLanguage;
    }
  }

  return "en";
}

export default i18n;
