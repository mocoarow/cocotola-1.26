import { useTranslation } from "react-i18next";

export function ProgressBar({ current, total }: { current: number; total: number }) {
  const { t } = useTranslation();
  const percent = total > 0 ? Math.round((current / total) * 100) : 0;

  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm text-muted-foreground">
        <span>{t("workbooks.study.progress", { current, total })}</span>
        <span>{percent}%</span>
      </div>
      <div className="h-2 w-full rounded-full bg-muted">
        <div
          className="h-2 rounded-full bg-primary transition-all duration-300"
          style={{ width: `${percent}%` }}
        />
      </div>
    </div>
  );
}
