import { BookOpenIcon } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher, useNavigate } from "react-router";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Button } from "~/components/ui/button";
import type { StudySummary } from "~/lib/api/study.server";

const STUDY_SIZE_STEP = 10;
const ABSOLUTE_MAX_STUDY_SIZE = 100;
const DEFAULT_STUDY_SIZE = 20;

type StartStudyDialogProps = {
  workbookId: string;
  triggerLabel: string;
  triggerClassName?: string;
};

// Dialog shown when the user clicks "Study". Loads the available
// new/review counts on demand (so the workbook list page itself does not
// need to fetch every workbook's summary up front) and lets the user pick
// how many questions this session should contain. The actual question
// selection (and the new/review mix) happens server-side once the user
// commits to /workbooks/:id/study?limit=N.
export function StartStudyDialog({
  workbookId,
  triggerLabel,
  triggerClassName,
}: StartStudyDialogProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const fetcher = useFetcher<StudySummary>();
  const [open, setOpen] = useState(false);
  const [selectedLimit, setSelectedLimit] = useState<number>(DEFAULT_STUDY_SIZE);

  // Load summary on first open. Re-opening reuses the cached fetcher data
  // (state is preserved across opens) so we only refetch if the user
  // explicitly retries.
  useEffect(() => {
    if (open && fetcher.state === "idle" && fetcher.data === undefined) {
      fetcher.load(`/workbooks/${workbookId}/study-summary`);
    }
  }, [open, fetcher, workbookId]);

  const summary = fetcher.data;
  const isLoading = fetcher.state === "loading" || (open && summary === undefined);
  const totalAvailable = summary?.totalDue ?? 0;
  const maxAllowed = Math.min(ABSOLUTE_MAX_STUDY_SIZE, totalAvailable);

  const sizeOptions = useMemo(() => {
    if (maxAllowed <= 0) return [] as { value: number; label: string }[];
    const options: { value: number; label: string }[] = [];
    for (let v = STUDY_SIZE_STEP; v <= maxAllowed; v += STUDY_SIZE_STEP) {
      options.push({ value: v, label: String(v) });
    }
    const lastStep = options.at(-1)?.value ?? 0;
    if (
      totalAvailable > 0 &&
      totalAvailable < ABSOLUTE_MAX_STUDY_SIZE &&
      totalAvailable !== lastStep
    ) {
      options.push({
        value: totalAvailable,
        label: t("workbooks.studyDialog.allButton", { count: totalAvailable }),
      });
    }
    return options;
  }, [maxAllowed, totalAvailable, t]);

  const effectiveLimit = useMemo(() => {
    if (sizeOptions.length === 0) return 0;
    const match = sizeOptions.find((o) => o.value === selectedLimit);
    if (match) return match.value;
    return sizeOptions.at(-1)?.value ?? 0;
  }, [sizeOptions, selectedLimit]);
  const canStart = !isLoading && totalAvailable > 0 && effectiveLimit > 0;

  function handleStart() {
    if (!canStart) return;
    navigate(`/workbooks/${workbookId}/study?limit=${effectiveLimit}`);
    setOpen(false);
  }

  return (
    <>
      <Button
        size="sm"
        className={triggerClassName}
        nativeButton={false}
        onClick={() => setOpen(true)}
      >
        <BookOpenIcon data-icon="inline-start" className="size-3.5" />
        <span>{triggerLabel}</span>
      </Button>

      <AlertDialog
        open={open}
        onOpenChange={(nextOpen) => {
          if (!nextOpen) setOpen(false);
        }}
      >
        <AlertDialogContent>
          <AlertDialogTitle>{t("workbooks.studyDialog.title")}</AlertDialogTitle>
          <AlertDialogDescription>{t("workbooks.studyDialog.description")}</AlertDialogDescription>

          {isLoading ? (
            <p className="mt-4 text-sm text-muted-foreground">
              {t("workbooks.studyDialog.loading")}
            </p>
          ) : summary === undefined ? (
            <p className="mt-4 text-sm text-destructive">{t("workbooks.studyDialog.loadError")}</p>
          ) : (
            <div className="mt-4 space-y-4">
              <dl className="grid grid-cols-2 gap-3 text-sm">
                <div className="rounded-md border p-3">
                  <dt className="text-xs text-muted-foreground">
                    {t("workbooks.studyDialog.reviewLabel")}
                  </dt>
                  <dd className="mt-1 text-lg font-semibold">{summary.reviewCount}</dd>
                </div>
                <div className="rounded-md border p-3">
                  <dt className="text-xs text-muted-foreground">
                    {t("workbooks.studyDialog.newLabel")}
                  </dt>
                  <dd className="mt-1 text-lg font-semibold">{summary.newCount}</dd>
                </div>
              </dl>
              <p className="text-xs text-muted-foreground">
                {t("workbooks.studyDialog.ratioHint", {
                  numerator: summary.reviewRatioNumerator,
                  denominator: summary.reviewRatioDenominator,
                })}
              </p>

              {totalAvailable === 0 ? (
                <p className="text-sm text-muted-foreground">
                  {t("workbooks.studyDialog.noneAvailable")}
                </p>
              ) : (
                <div>
                  <label htmlFor="study-size-select" className="mb-2 block text-sm font-medium">
                    {t("workbooks.studyDialog.sizeLabel")}
                  </label>
                  <select
                    id="study-size-select"
                    value={effectiveLimit}
                    onChange={(e) => {
                      const next = Number(e.target.value);
                      if (Number.isFinite(next)) setSelectedLimit(next);
                    }}
                    className="h-9 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
                  >
                    {sizeOptions.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                  <p className="mt-2 text-xs text-muted-foreground">
                    {t("workbooks.studyDialog.selected", { count: effectiveLimit })}
                  </p>
                </div>
              )}
            </div>
          )}

          <AlertDialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              {t("common.cancel")}
            </Button>
            <Button onClick={handleStart} disabled={!canStart}>
              {t("workbooks.studyDialog.start")}
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
