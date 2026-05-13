import { ArrowLeftIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Link, useLoaderData } from "react-router";
import { listStudyRecords, type StudyRecord } from "~/lib/api/study.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId.debug-study";

// Debug-only route. Production builds make this loader throw so the page is
// effectively unreachable; the navigation button in `workbooks.$workbookId`
// is also hidden behind `import.meta.env.DEV` so production users can't see
// the URL in the UI to begin with.
export async function loader({ request, params }: Route.LoaderArgs) {
  if (process.env.NODE_ENV === "production") {
    throw new Response("Not Found", { status: 404 });
  }
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const { records } = await listStudyRecords(accessToken, workbookId);
  return { workbookId, records };
}

function formatDate(value: string): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  // ISO with seconds, no millis — easier to compare during debugging.
  return d.toISOString().replace(/\.\d{3}Z$/, "Z");
}

export default function DebugStudyRecords() {
  const { workbookId, records } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  return (
    <div>
      <Link
        to={`/workbooks/${workbookId}`}
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeftIcon className="size-3.5" />
        {t("workbooks.debugStudy.back")}
      </Link>

      <h1 className="text-2xl font-bold">{t("workbooks.debugStudy.title")}</h1>
      <p className="mt-1 mb-4 text-sm text-muted-foreground">
        {t("workbooks.debugStudy.description")}
      </p>

      {records.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-12">
          <p className="text-sm text-muted-foreground">{t("workbooks.debugStudy.empty")}</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-lg border bg-card">
          <table className="min-w-full text-sm">
            <thead className="bg-muted/50 text-left text-xs uppercase text-muted-foreground">
              <tr>
                <th className="px-3 py-2">{t("workbooks.debugStudy.questionId")}</th>
                <th className="px-3 py-2">{t("workbooks.debugStudy.consecutiveCorrect")}</th>
                <th className="px-3 py-2">{t("workbooks.debugStudy.totalCorrect")}</th>
                <th className="px-3 py-2">{t("workbooks.debugStudy.totalIncorrect")}</th>
                <th className="px-3 py-2">{t("workbooks.debugStudy.lastAnsweredAt")}</th>
                <th className="px-3 py-2">{t("workbooks.debugStudy.nextDueAt")}</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {records.map((r: StudyRecord) => (
                <tr key={r.questionId}>
                  <td className="whitespace-nowrap px-3 py-2 font-mono text-xs">{r.questionId}</td>
                  <td className="whitespace-nowrap px-3 py-2 tabular-nums">
                    {r.consecutiveCorrect}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 tabular-nums">{r.totalCorrect}</td>
                  <td className="whitespace-nowrap px-3 py-2 tabular-nums">{r.totalIncorrect}</td>
                  <td className="whitespace-nowrap px-3 py-2 font-mono text-xs">
                    {formatDate(r.lastAnsweredAt)}
                  </td>
                  <td className="whitespace-nowrap px-3 py-2 font-mono text-xs">
                    {formatDate(r.nextDueAt)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
