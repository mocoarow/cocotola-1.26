import { TargetIcon } from "lucide-react";
import { useTranslation } from "react-i18next";

function clampRatio(today: number, goal: number): number {
  if (goal <= 0) return 0;
  return Math.min(1, Math.max(0, today / goal));
}

export function DailyGoalCard({
  todayCount,
  dailyGoal,
}: {
  todayCount: number;
  dailyGoal: number;
}) {
  const { t } = useTranslation();
  const ratio = clampRatio(todayCount, dailyGoal);
  const percent = Math.round(ratio * 100);
  const reached = todayCount >= dailyGoal && dailyGoal > 0;

  return (
    <section
      aria-label={t("dashboard.goal.label")}
      className="rounded-md border bg-card p-4"
      data-testid="daily-goal-card"
    >
      <div className="flex items-center gap-2 text-sm font-semibold">
        <TargetIcon className="size-4 text-emerald-500" aria-hidden="true" />
        <span>{t("dashboard.goal.heading")}</span>
      </div>
      <p
        className="mt-2 text-3xl font-bold"
        data-testid="daily-goal-progress"
        data-reached={reached}
      >
        {t("dashboard.goal.progress", { today: todayCount, goal: dailyGoal })}
      </p>
      <div
        role="progressbar"
        aria-valuenow={percent}
        aria-valuemin={0}
        aria-valuemax={100}
        className="mt-3 h-2 w-full rounded-full bg-muted"
      >
        <div
          className={`h-2 rounded-full transition-all duration-300 ${
            reached ? "bg-emerald-500" : "bg-primary"
          }`}
          style={{ width: `${percent}%` }}
        />
      </div>
      <p className="mt-2 text-xs text-muted-foreground">
        {reached
          ? t("dashboard.goal.reached")
          : t("dashboard.goal.remaining", { count: Math.max(0, dailyGoal - todayCount) })}
      </p>
    </section>
  );
}
