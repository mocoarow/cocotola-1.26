import { CheckIcon, GlobeIcon, PlusIcon } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useFetcher, useLoaderData } from "react-router";
import { useConfirm } from "~/components/confirm-dialog-provider";
import { StartStudyDialog } from "~/components/study/start-study-dialog";
import { Button } from "~/components/ui/button";
import {
  listPublicWorkbooks,
  listSharedWorkbooks,
  type PublicWorkbook,
  shareWorkbook,
  unshareWorkbook,
} from "~/lib/api/sharing.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { formatDate } from "~/lib/format/date";
import type { Route } from "./+types/workbooks.public";

export async function loader({ request }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);

  const [workbooks, references] = await Promise.all([
    listPublicWorkbooks(accessToken),
    listSharedWorkbooks(accessToken),
  ]);

  return { workbooks, references };
}

async function callTolerantly(
  fn: () => Promise<unknown>,
  tolerableStatuses: readonly number[],
): Promise<{ tolerated: boolean }> {
  try {
    await fn();
    return { tolerated: false };
  } catch (error) {
    if (error instanceof Response && tolerableStatuses.includes(error.status)) {
      return { tolerated: true };
    }
    throw error;
  }
}

export async function action({ request }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "share") {
    const workbookId = formData.get("workbookId");
    if (typeof workbookId !== "string" || !workbookId) {
      throw new Response("workbookId is required", { status: 400 });
    }
    // 409 means the user already shared this workbook (double-click or
    // concurrent tab). Treat as success so the next loader pass re-syncs.
    const { tolerated } = await callTolerantly(() => shareWorkbook(accessToken, workbookId), [409]);
    return { ok: true, alreadyShared: tolerated };
  }

  if (intent === "unshare") {
    const referenceId = formData.get("referenceId");
    if (typeof referenceId !== "string" || !referenceId) {
      throw new Response("referenceId is required", { status: 400 });
    }
    // 404 means the reference was already removed (double-click or
    // concurrent tab). Treat as success so the next loader pass re-syncs.
    const { tolerated } = await callTolerantly(
      () => unshareWorkbook(accessToken, referenceId),
      [404],
    );
    return { ok: true, alreadyRemoved: tolerated };
  }

  throw new Response("Unknown intent", { status: 400 });
}

function PublicWorkbookCard({
  workbook,
  referenceId,
}: {
  workbook: PublicWorkbook;
  referenceId: string | undefined;
}) {
  const fetcher = useFetcher();
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();
  const isShared = referenceId !== undefined;
  const isSubmitting = fetcher.state !== "idle";

  async function handleUnshare(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!referenceId) return;
    const confirmed = await confirm({
      title: t("workbooks.public.removeConfirmTitle"),
      description: t("workbooks.public.removeConfirm", { title: workbook.title }),
      confirmLabel: t("workbooks.public.remove"),
    });
    if (confirmed) {
      fetcher.submit({ intent: "unshare", referenceId }, { method: "post" });
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
        <span className="shrink-0 rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
          {t(`languages.${workbook.language}`, { defaultValue: workbook.language })}
        </span>
      </div>

      {workbook.description && (
        <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">{workbook.description}</p>
      )}

      <p className="mb-4 text-xs text-muted-foreground">
        {t("workbooks.public.published")} {formatDate(workbook.createdAt, i18n.language)}
      </p>

      <div className="flex items-center gap-2">
        <StartStudyDialog
          workbookId={workbook.workbookId}
          triggerLabel={t("workbooks.public.study")}
          triggerClassName="flex-1"
        />
        {isShared ? (
          <fetcher.Form method="post" className="flex-1" onSubmit={handleUnshare}>
            <Button
              type="submit"
              variant="outline"
              size="sm"
              className="w-full"
              disabled={isSubmitting}
              aria-label={t("workbooks.public.removeAriaLabel", { title: workbook.title })}
              title={t("workbooks.public.remove")}
            >
              <CheckIcon data-icon="inline-start" className="size-3.5" />
              <span>{t("workbooks.public.added")}</span>
            </Button>
          </fetcher.Form>
        ) : (
          <fetcher.Form method="post" className="flex-1">
            <input type="hidden" name="intent" value="share" />
            <input type="hidden" name="workbookId" value={workbook.workbookId} />
            <Button
              type="submit"
              size="sm"
              className="w-full"
              disabled={isSubmitting}
              aria-label={t("workbooks.public.addAriaLabel", { title: workbook.title })}
            >
              <PlusIcon data-icon="inline-start" className="size-3.5" />
              <span>{t("workbooks.public.add")}</span>
            </Button>
          </fetcher.Form>
        )}
      </div>
    </div>
  );
}

export default function WorkbooksPublic() {
  const { workbooks, references } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  const referenceByWorkbookId = useMemo(
    () => new Map(references.map((ref) => [ref.workbookId, ref.referenceId] as const)),
    [references],
  );

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">{t("workbooks.public.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("workbooks.public.description")}</p>
      </div>

      {workbooks.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <GlobeIcon className="mb-4 size-12 text-muted-foreground/50" />
          <p className="text-lg font-medium text-muted-foreground">
            {t("workbooks.public.empty.title")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            {t("workbooks.public.empty.description")}
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {workbooks.map((workbook) => (
            <PublicWorkbookCard
              key={workbook.workbookId}
              workbook={workbook}
              referenceId={referenceByWorkbookId.get(workbook.workbookId)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
