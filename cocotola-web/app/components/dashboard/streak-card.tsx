import { FlameIcon } from "lucide-react";
import { useTranslation } from "react-i18next";

export function StreakCard({ current, longest }: { current: number; longest: number }) {
  const { t } = useTranslation();

  return (
    <section
      aria-label={t("dashboard.streak.label")}
      className="rounded-md border bg-card p-4"
      data-testid="streak-card"
    >
      <div className="flex items-center gap-2 text-sm font-semibold">
        <FlameIcon className="size-4 text-amber-500" aria-hidden="true" />
        <span>{t("dashboard.streak.heading")}</span>
      </div>
      <p className="mt-2 text-3xl font-bold">
        {t("dashboard.streak.currentValue", { count: current })}
      </p>
      <p className="text-xs text-muted-foreground" data-testid="streak-longest">
        {t("dashboard.streak.longest", { count: longest })}
      </p>
    </section>
  );
}
