import { useTranslation } from "react-i18next";

const LANG_KEYS = [
  { value: "ja", labelKey: "languages.ja" },
  { value: "en", labelKey: "languages.en" },
  { value: "it", labelKey: "languages.it" },
  { value: "fr", labelKey: "languages.fr" },
  { value: "de", labelKey: "languages.de" },
  { value: "es", labelKey: "languages.es" },
  { value: "zh", labelKey: "languages.zh" },
  { value: "ko", labelKey: "languages.ko" },
  { value: "pt", labelKey: "languages.pt" },
];

export const DEFAULT_SOURCE_LANG = "ja";
export const DEFAULT_TARGET_LANG = "en";

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
      {LANG_KEYS.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {t(opt.labelKey)}
        </option>
      ))}
    </select>
  );
}
