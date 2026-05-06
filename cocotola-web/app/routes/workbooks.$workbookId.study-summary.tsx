import { getStudySummary } from "~/lib/api/study.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId.study-summary";

// Resource route — no default export. Used by the study-size picker dialog
// on the workbook list to render available counts (new / review) and the
// server-side review/new ratio without downloading any questions.
export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const url = new URL(request.url);
  const practice = url.searchParams.get("practice") === "true";
  const summary = await getStudySummary(accessToken, workbookId, practice);
  return summary;
}
