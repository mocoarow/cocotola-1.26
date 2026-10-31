import { XIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useFetcher } from "react-router";
import { useConfirm } from "~/components/confirm-dialog-provider";
import { StartStudyDialog } from "~/components/study/start-study-dialog";
import { Button } from "~/components/ui/button";
import type { SharedWorkbook } from "~/lib/api/sharing.server";
import { formatDate } from "~/lib/format/date";

export function AddedWorkbookCard({ workbook }: { workbook: SharedWorkbook }) {
  const fetcher = useFetcher();
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();
  const isSubmitting = fetcher.state !== "idle";

  async function handleRemove() {
    const confirmed = await confirm({
      title: t("workbooks.public.removeConfirmTitle"),
      description: t("workbooks.public.removeConfirm", { title: workbook.title }),
      confirmLabel: t("workbooks.public.remove"),
    });
    if (confirmed) {
      fetcher.submit({ intent: "unshare", referenceId: workbook.referenceId }, { method: "post" });
    }
  }

  return (
    <div
      className={`group rounded-lg border bg-card p-5 shadow-sm transition-shadow hover:shadow-md ${
        isSubmitting ? "opacity-50" : ""
      }`}
    >
      <div className="mb-3 flex items-start justify-between gap-2">
        <h3 className="text-base font-semibold leading-tight">{workbook.title}</h3>
        <div className="flex shrink-0 gap-1">
          <span className="rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
            {t("workbooks.index.addedBadge")}
          </span>
          <span className="rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
            {t(`languages.${workbook.language}`, { defaultValue: workbook.language })}
          </span>
        </div>
      </div>

      {workbook.description && (
        <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">{workbook.description}</p>
      )}

      <p className="mb-4 text-xs text-muted-foreground">
        {t("workbooks.index.addedAt")} {formatDate(workbook.addedAt, i18n.language)}
      </p>

      <div className="flex items-center gap-2">
        <StartStudyDialog
          workbookId={workbook.workbookId}
          triggerLabel={t("workbooks.index.study")}
          triggerClassName="flex-1"
        />
        <Button
          variant="outline"
          size="sm"
          className="flex-1"
          disabled={isSubmitting}
          onClick={handleRemove}
          aria-label={t("workbooks.public.removeAriaLabel", { title: workbook.title })}
        >
          <XIcon data-icon="inline-start" className="size-3.5" />
          <span>{t("workbooks.public.remove")}</span>
        </Button>
      </div>
    </div>
  );
}
