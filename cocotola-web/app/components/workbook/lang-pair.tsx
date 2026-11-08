import { useState } from "react";
import { useTranslation } from "react-i18next";
import { DEFAULT_SOURCE_LANG, DEFAULT_TARGET_LANG, LangSelect } from "./lang-select";

export const SAME_LANGUAGE_ERROR_KEY = "workbooks.addQuestion.errors.sameLanguage";

export type LangPair = {
  sourceLang: string;
  setSourceLang: (lang: string) => void;
  targetLang: string;
  setTargetLang: (lang: string) => void;
  sameLanguage: boolean;
};

export function useLangPair(savedSourceLang?: string, savedTargetLang?: string): LangPair {
  const [sourceLang, setSourceLang] = useState(savedSourceLang ?? DEFAULT_SOURCE_LANG);
  const [targetLang, setTargetLang] = useState(savedTargetLang ?? DEFAULT_TARGET_LANG);
  return {
    sourceLang,
    setSourceLang,
    targetLang,
    setTargetLang,
    sameLanguage: sourceLang === targetLang,
  };
}

export function SourceLangField({ id, langs }: { id: string; langs: LangPair }) {
  const { t } = useTranslation();
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {t("workbooks.addQuestion.sourceLang")}
      </label>
      <LangSelect
        id={id}
        name="sourceLang"
        value={langs.sourceLang}
        onChange={langs.setSourceLang}
      />
    </div>
  );
}

export function TargetLangField({ id, langs }: { id: string; langs: LangPair }) {
  const { t } = useTranslation();
  const errorId = `${id}-error`;
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {t("workbooks.addQuestion.targetLang")}
      </label>
      <LangSelect
        id={id}
        name="targetLang"
        value={langs.targetLang}
        onChange={langs.setTargetLang}
        aria-invalid={langs.sameLanguage}
        aria-describedby={langs.sameLanguage ? errorId : undefined}
      />
      <p id={errorId} aria-live="polite" className="text-xs text-destructive">
        {langs.sameLanguage && t(SAME_LANGUAGE_ERROR_KEY)}
      </p>
    </div>
  );
}
