import { useTranslation } from "react-i18next";
import {
  CONTENT_LANGUAGES,
  type ContentLanguage,
  DEFAULT_CONTENT_LANGUAGE,
} from "~/lib/workbook/content-language";

export const DEFAULT_SOURCE_LANG: ContentLanguage = DEFAULT_CONTENT_LANGUAGE;
export const DEFAULT_TARGET_LANG: ContentLanguage = "en";

export function LangSelect({
  id,
  name,
  value,
  onChange,
  "aria-invalid": ariaInvalid,
  "aria-describedby": ariaDescribedBy,
}: {
  id: string;
  name: string;
  value: string;
  onChange: (value: string) => void;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
}) {
  const { t } = useTranslation();
  return (
    <select
      id={id}
      name={name}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      aria-invalid={ariaInvalid}
      aria-describedby={ariaDescribedBy}
      className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
    >
      {CONTENT_LANGUAGES.map((lang) => (
        <option key={lang} value={lang}>
          {t(`languages.${lang}`)}
        </option>
      ))}
    </select>
  );
}
