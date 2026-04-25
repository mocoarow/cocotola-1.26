import { CheckCircleIcon, TrophyIcon, XCircleIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { Button } from "~/components/ui/button";

type StudyResultProps = {
  correctCount: number;
  incorrectCount: number;
  workbookId: string;
};

export function StudyResult({ correctCount, incorrectCount, workbookId }: StudyResultProps) {
  const { t } = useTranslation();
  const total = correctCount + incorrectCount;
  const percent = total > 0 ? Math.round((correctCount / total) * 100) : 0;

  return (
    <div className="flex flex-col items-center justify-center py-12">
      <TrophyIcon className="mb-4 size-16 text-yellow-500" />
      <h2 className="mb-2 text-2xl font-bold">{t("workbooks.study.result.title")}</h2>
      <p className="mb-8 text-lg text-muted-foreground">
        {t("workbooks.study.result.score", { percent })}
      </p>

      <div className="mb-8 grid grid-cols-3 gap-8 text-center">
        <div>
          <div className="flex items-center justify-center gap-1">
            <CheckCircleIcon className="size-5 text-green-600" />
            <span className="text-2xl font-bold text-green-600">{correctCount}</span>
          </div>
          <p className="text-sm text-muted-foreground">{t("workbooks.study.result.correct")}</p>
        </div>
        <div>
          <div className="flex items-center justify-center gap-1">
            <XCircleIcon className="size-5 text-red-600" />
            <span className="text-2xl font-bold text-red-600">{incorrectCount}</span>
          </div>
          <p className="text-sm text-muted-foreground">{t("workbooks.study.result.incorrect")}</p>
        </div>
        <div>
          <span className="text-2xl font-bold">{total}</span>
          <p className="text-sm text-muted-foreground">{t("workbooks.study.result.total")}</p>
        </div>
      </div>

      <Button nativeButton={false} render={<Link to={`/workbooks/${workbookId}`} />}>
        {t("workbooks.study.backToWorkbook")}
      </Button>
    </div>
  );
}
