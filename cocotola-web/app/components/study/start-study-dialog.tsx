import { BookOpenIcon } from "lucide-react";
import { useEffect, useState } from "react";
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

const STUDY_SIZE_PRESETS = [10, 20, 50] as const;
const ABSOLUTE_MAX_STUDY_SIZE = 100;

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
  const [selectedLimit, setSelectedLimit] = useState<number>(STUDY_SIZE_PRESETS[1]);

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
  const maxAllowed = Math.min(ABSOLUTE_MAX_STUDY_SIZE, Math.max(1, totalAvailable));
  const effectiveLimit = Math.min(selectedLimit, maxAllowed);
  const canStart = !isLoading && totalAvailable > 0;

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
          <AlertDialogDescription>
            {t("workbooks.studyDialog.description")}
          </AlertDialogDescription>

          {isLoading ? (
            <p className="mt-4 text-sm text-muted-foreground">
              {t("workbooks.studyDialog.loading")}
            </p>
          ) : summary === undefined ? (
            <p className="mt-4 text-sm text-destructive">
              {t("workbooks.studyDialog.loadError")}
            </p>
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
                  <p className="mb-2 text-sm font-medium">
                    {t("workbooks.studyDialog.sizeLabel")}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {STUDY_SIZE_PRESETS.map((preset) => {
                      const disabled = preset > maxAllowed && preset !== STUDY_SIZE_PRESETS[0];
                      const active = effectiveLimit === Math.min(preset, maxAllowed);
                      return (
                        <Button
                          key={preset}
                          type="button"
                          size="sm"
                          variant={active ? "default" : "outline"}
                          disabled={disabled}
                          onClick={() => setSelectedLimit(preset)}
                        >
                          {preset}
                        </Button>
                      );
                    })}
                    {totalAvailable < STUDY_SIZE_PRESETS[STUDY_SIZE_PRESETS.length - 1] && (
                      <Button
                        type="button"
                        size="sm"
                        variant={effectiveLimit === totalAvailable ? "default" : "outline"}
                        onClick={() => setSelectedLimit(totalAvailable)}
                      >
                        {t("workbooks.studyDialog.allButton", { count: totalAvailable })}
                      </Button>
                    )}
                  </div>
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
