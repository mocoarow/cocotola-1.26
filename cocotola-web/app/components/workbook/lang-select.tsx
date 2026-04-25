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

export function LangSelect({
  id,
  name,
  defaultValue,
}: {
  id: string;
  name: string;
  defaultValue: string;
}) {
  const { t } = useTranslation();
  return (
    <select
      id={id}
      name={name}
      defaultValue={defaultValue}
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
