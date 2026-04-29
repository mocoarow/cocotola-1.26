import { BookOpenIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useFetcher, useLoaderData } from "react-router";
import { useConfirm } from "~/components/confirm-dialog-provider";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "~/components/ui/sheet";
import { findPrivateSpace } from "~/lib/api/space.server";
import {
  createWorkbook,
  deleteWorkbook,
  listWorkbooks,
  type Workbook,
} from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { formatDate } from "~/lib/format/date";
import type { Route } from "./+types/workbooks.index";

export async function loader({ request }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);

  const privateSpace = await findPrivateSpace(accessToken);
  if (!privateSpace) {
    return { workbooks: [], spaceId: null };
  }

  const workbooks = await listWorkbooks(accessToken, privateSpace.spaceId);
  return { workbooks, spaceId: privateSpace.spaceId };
}

export async function action({ request }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "create") {
    const privateSpace = await findPrivateSpace(accessToken);
    if (!privateSpace) {
      throw new Response("No private space found", { status: 400 });
    }
    const title = String(formData.get("title") ?? "").trim();
    const description = String(formData.get("description") ?? "").trim();

    if (!title) {
      return { ok: false, errorKey: "workbooks.index.errors.titleRequired" };
    }
    if (title.length > 200) {
      return { ok: false, errorKey: "workbooks.index.errors.titleTooLong" };
    }
    if (description.length > 1000) {
      return { ok: false, errorKey: "workbooks.index.errors.descriptionTooLong" };
    }

    await createWorkbook(accessToken, {
      spaceId: privateSpace.spaceId,
      title,
      description,
      visibility: "private",
    });
  }

  if (intent === "delete") {
    const workbookId = formData.get("workbookId");
    if (typeof workbookId === "string") {
      await deleteWorkbook(accessToken, workbookId);
    }
  }

  return { ok: true };
}

function WorkbookCard({ workbook }: { workbook: Workbook }) {
  const fetcher = useFetcher();
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();
  const isDeleting = fetcher.state !== "idle";

  return (
    <div
      className={`group rounded-lg border bg-card p-5 shadow-sm transition-shadow hover:shadow-md ${
        isDeleting ? "opacity-50" : ""
      }`}
    >
      <div className="mb-3 flex items-start justify-between gap-2">
        <h3 className="text-base font-semibold leading-tight">{workbook.title}</h3>
        <span
          className={`shrink-0 rounded-full px-2 py-0.5 text-[11px] font-medium ${
            workbook.visibility === "public"
              ? "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              : "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400"
          }`}
        >
          {workbook.visibility}
        </span>
      </div>

      {workbook.description && (
        <p className="mb-4 line-clamp-2 text-sm text-muted-foreground">{workbook.description}</p>
      )}

      <p className="mb-4 text-xs text-muted-foreground">
        {t("workbooks.index.updated")} {formatDate(workbook.updatedAt, i18n.language)}
      </p>

      <div className="flex items-center gap-2">
        <Button
          size="sm"
          className="flex-1"
          nativeButton={false}
          render={<Link to={`/workbooks/${workbook.workbookId}/study`} />}
        >
          <BookOpenIcon data-icon="inline-start" className="size-3.5" />
          <span>{t("workbooks.index.study")}</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="flex-1"
          nativeButton={false}
          render={<Link to={`/workbooks/${workbook.workbookId}`} />}
        >
          <PencilIcon data-icon="inline-start" className="size-3.5" />
          <span>{t("workbooks.index.edit")}</span>
        </Button>
        <Button
          variant="destructive"
          size="icon-sm"
          disabled={isDeleting}
          onClick={async () => {
            const confirmed = await confirm({
              title: t("workbooks.index.deleteConfirmTitle"),
              description: t("workbooks.index.deleteConfirm", { title: workbook.title }),
              confirmLabel: t("common.delete"),
            });
            if (confirmed) {
              fetcher.submit(
                { intent: "delete", workbookId: workbook.workbookId },
                { method: "post" },
              );
            }
          }}
        >
          <Trash2Icon className="size-3.5" />
          <span className="sr-only">{t("common.delete")}</span>
        </Button>
      </div>
    </div>
  );
}

function CreateWorkbookButton() {
  const fetcher = useFetcher();
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const isSubmitting = fetcher.state !== "idle";

  if (fetcher.data?.ok && open) {
    setOpen(false);
    setFormKey((k) => k + 1);
  }

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger
        render={
          <Button size="sm">
            <PlusIcon data-icon="inline-start" className="size-4" />
            <span>{t("workbooks.index.newWorkbook")}</span>
          </Button>
        }
      />
      <SheetContent side="right">
        <SheetHeader>
          <SheetTitle>{t("workbooks.index.createTitle")}</SheetTitle>
          <SheetDescription>{t("workbooks.index.createDescription")}</SheetDescription>
        </SheetHeader>
        <fetcher.Form key={formKey} method="post" className="flex flex-1 flex-col gap-4 px-4">
          <input type="hidden" name="intent" value="create" />
          <div className="space-y-1.5">
            <label htmlFor="title" className="text-sm font-medium">
              {t("workbooks.index.titleLabel")} <span className="text-destructive">*</span>
            </label>
            <Input
              id="title"
              name="title"
              required
              maxLength={200}
              placeholder={t("workbooks.index.titlePlaceholder")}
            />
          </div>
          <div className="space-y-1.5">
            <label htmlFor="description" className="text-sm font-medium">
              {t("workbooks.index.descriptionLabel")}
            </label>
            <Input
              id="description"
              name="description"
              maxLength={1000}
              placeholder={t("workbooks.index.descriptionPlaceholder")}
            />
          </div>
          {fetcher.data && !fetcher.data.ok && "errorKey" in fetcher.data && (
            <p className="text-sm text-destructive">{t(fetcher.data.errorKey)}</p>
          )}
          <SheetFooter>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? t("workbooks.index.creating") : t("common.create")}
            </Button>
          </SheetFooter>
        </fetcher.Form>
      </SheetContent>
    </Sheet>
  );
}

export default function WorkbooksIndex() {
  const { workbooks } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{t("workbooks.index.title")}</h1>
          <p className="mt-1 text-sm text-muted-foreground">{t("workbooks.index.description")}</p>
        </div>
        <CreateWorkbookButton />
      </div>

      {workbooks.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <BookOpenIcon className="mb-4 size-12 text-muted-foreground/50" />
          <p className="text-lg font-medium text-muted-foreground">
            {t("workbooks.index.empty.title")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            {t("workbooks.index.empty.description")}
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {workbooks.map((workbook) => (
            <WorkbookCard key={workbook.workbookId} workbook={workbook} />
          ))}
        </div>
      )}
    </div>
  );
}
