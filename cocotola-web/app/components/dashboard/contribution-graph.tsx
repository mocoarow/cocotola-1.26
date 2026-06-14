import { useMemo } from "react";
import { useTranslation } from "react-i18next";

import type { DashboardDailyItem } from "~/lib/api/dashboard.server";

/**
 * Five-step Tailwind palette mirroring the GitHub contribution graph.
 * Index 0 = no activity; index 4 = the day's `answeredCount` reached
 * `COLOR_STEPS[4]` or more.
 */
const COLOR_STEPS = [0, 1, 5, 10, 20];

const TONE_CLASSES = [
  "bg-muted",
  "bg-emerald-200 dark:bg-emerald-900",
  "bg-emerald-400 dark:bg-emerald-700",
  "bg-emerald-500 dark:bg-emerald-600",
  "bg-emerald-700 dark:bg-emerald-400",
];

function toneIndex(count: number): number {
  for (let i = COLOR_STEPS.length - 1; i >= 0; i--) {
    if (count >= COLOR_STEPS[i]) return i;
  }
  return 0;
}

const DAY_LABEL_ROWS = [1, 3, 5]; // Mon / Wed / Fri

type Cell = {
  readonly key: string;
  readonly date: string;
  readonly answeredCount: number;
  readonly correctCount: number;
};

type Week = {
  readonly key: string;
  readonly cells: readonly Cell[];
};

type GraphMatrix = {
  readonly weeks: readonly Week[];
  readonly monthLabels: readonly { weekKey: string; month: string }[];
};

// Project policy bans array.push and mid-iteration state mutation, so
// each stage of buildMatrix takes an input array and returns a fresh one
// (padding cells, week chunks, month labels). The dataset is bounded by
// the dashboard window (max 730 days = ~104 weeks) so the extra GC
// pressure from intermediate arrays is irrelevant compared to the
// readability win of a pipeline of pure functions.
function buildMatrix(items: readonly DashboardDailyItem[], locale: string): GraphMatrix {
  if (items.length === 0) return { weeks: [], monthLabels: [] };

  const padded = padToWeekGrid(items);
  const weeks = chunkIntoWeeks(padded);
  const monthLabels = collectMonthLabels(weeks, locale);

  return { weeks, monthLabels };
}

/**
 * Builds a UTC Date for the supplied YYYY-MM-DD key by splitting the
 * components and feeding them to Date.UTC, instead of relying on
 * `new Date("YYYY-MM-DDT00:00:00Z")` parsing. Both are equivalent on
 * spec-compliant runtimes, but the explicit construction makes the
 * "no timezone math, just a calendar point in UTC" intent obvious to
 * the reader and avoids any runtime-specific ISO-8601 parsing quirks
 * that historically shifted by a day around offset boundaries.
 */
function utcDateFromDateKey(dateKey: string): Date {
  const [year, month, day] = dateKey.split("-").map(Number);
  return new Date(Date.UTC(year, month - 1, day));
}

function padToWeekGrid(items: readonly DashboardDailyItem[]): readonly Cell[] {
  const offsetToSunday = utcDateFromDateKey(items[0].date).getUTCDay();

  const leadingPad: Cell[] = Array.from({ length: offsetToSunday }, (_, i) => ({
    key: `pad-${items[0].date}-${i}`,
    date: "",
    answeredCount: 0,
    correctCount: 0,
  }));

  const dataCells: Cell[] = items.map((item) => ({
    key: `cell-${item.date}`,
    date: item.date,
    answeredCount: item.answeredCount,
    correctCount: item.correctCount,
  }));

  return [...leadingPad, ...dataCells];
}

function chunkIntoWeeks(padded: readonly Cell[]): readonly Week[] {
  const weekCount = Math.ceil(padded.length / 7);
  return Array.from({ length: weekCount }, (_, weekIndex) => {
    const start = weekIndex * 7;
    const cells = padded.slice(start, start + 7);
    const anchor = cells.find((c) => c.date !== "");
    return {
      key: anchor ? `week-${anchor.date}` : `week-pad-${start}`,
      cells,
    };
  });
}

function collectMonthLabels(
  weeks: readonly Week[],
  locale: string,
): readonly { weekKey: string; month: string }[] {
  // timeZone: "UTC" pairs with utcDateFromDateKey above so the formatter
  // reports the same month the input string carries. Leaving it on the
  // runtime default would let a UTC+12 server return "July" for
  // "2025-06-30" because the UTC instant resolves to July 1 in local
  // time — a one-month shift that drifts the entire row of month labels.
  const monthFormatter = new Intl.DateTimeFormat(locale, {
    month: "short",
    timeZone: "UTC",
  });

  // Each week resolves to its first non-padding cell's month, or null when
  // the week is entirely padding. reduce then keeps only the transitions
  // (month different from the previous emitted one) so consecutive weeks
  // in the same month do not produce duplicate labels.
  const weekMonths = weeks.map((week) => {
    const firstReal = week.cells.find((c) => c.date !== "");
    if (!firstReal) return null;
    return {
      weekKey: week.key,
      month: monthFormatter.format(utcDateFromDateKey(firstReal.date)),
    };
  });

  return weekMonths.reduce<readonly { weekKey: string; month: string }[]>((labels, entry) => {
    if (entry === null) return labels;
    const previousMonth = labels.length === 0 ? "" : labels[labels.length - 1].month;
    if (entry.month === previousMonth) return labels;
    return [...labels, entry];
  }, []);
}

export function ContributionGraph({
  items,
  from,
  to,
  locale,
}: {
  items: readonly DashboardDailyItem[];
  from: string;
  to: string;
  locale: string;
}) {
  const { t } = useTranslation();
  const matrix = useMemo(() => buildMatrix(items, locale), [items, locale]);

  if (matrix.weeks.length === 0) {
    return (
      <div
        className="rounded-md border bg-card p-6 text-sm text-muted-foreground"
        data-testid="contribution-graph-empty"
      >
        {t("dashboard.empty")}
      </div>
    );
  }

  return (
    <section
      aria-label={t("dashboard.graph.label")}
      className="space-y-2 rounded-md border bg-card p-4"
      data-testid="contribution-graph"
    >
      <div className="flex items-baseline justify-between gap-2 text-sm">
        <h2 className="font-semibold">{t("dashboard.graph.heading")}</h2>
        <span className="text-xs text-muted-foreground">
          {t("dashboard.graph.range", { from, to })}
        </span>
      </div>

      <div className="flex gap-2 overflow-x-auto pb-2">
        <div className="flex flex-col justify-between pt-4 text-[10px] text-muted-foreground">
          {DAY_LABEL_ROWS.map((row) => (
            <span key={row} className="h-3">
              {t(`dashboard.graph.day.${row}`)}
            </span>
          ))}
        </div>

        <div className="flex flex-col">
          <div className="flex h-4 text-[10px] text-muted-foreground" aria-hidden="true">
            {matrix.weeks.map((week) => {
              const label = matrix.monthLabels.find((m) => m.weekKey === week.key);
              return (
                <span key={week.key} className="w-3 mr-[2px]">
                  {label?.month}
                </span>
              );
            })}
          </div>

          <div className="flex">
            {matrix.weeks.map((week) => (
              <div key={week.key} className="mr-[2px] flex flex-col gap-[2px]">
                {week.cells.map((cell) => {
                  if (cell.date === "") {
                    return <div key={cell.key} className="h-3 w-3" aria-hidden="true" />;
                  }
                  const tone = toneIndex(cell.answeredCount);
                  const toneClass = TONE_CLASSES[tone];
                  const title = t("dashboard.graph.cellTitle", {
                    date: cell.date,
                    count: cell.answeredCount,
                    correct: cell.correctCount,
                  });
                  return (
                    <div
                      key={cell.key}
                      title={title}
                      role="img"
                      aria-label={title}
                      data-testid="contribution-cell"
                      data-date={cell.date}
                      data-count={cell.answeredCount}
                      data-tone={tone}
                      className={`h-3 w-3 rounded-sm ${toneClass}`}
                    />
                  );
                })}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-end gap-2 text-[10px] text-muted-foreground">
        <span>{t("dashboard.graph.legendLess")}</span>
        {TONE_CLASSES.map((tone, i) => (
          <span
            key={tone}
            className={`h-3 w-3 rounded-sm ${tone}`}
            aria-hidden="true"
            data-tone={i}
          />
        ))}
        <span>{t("dashboard.graph.legendMore")}</span>
      </div>
    </section>
  );
}
