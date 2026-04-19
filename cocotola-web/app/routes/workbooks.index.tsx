import { useState } from "react";
import { BookOpenIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { Link, useFetcher, useLoaderData } from "react-router";
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
import { createWorkbook, deleteWorkbook, listWorkbooks, type Workbook } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
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
    const visibility = formData.get("visibility") === "public" ? "public" as const : "private" as const;

    if (!title) {
      return { ok: false, error: "Title is required" };
    }
    if (title.length > 200) {
      return { ok: false, error: "Title must be 200 characters or less" };
    }
    if (description.length > 1000) {
      return { ok: false, error: "Description must be 1000 characters or less" };
    }

    await createWorkbook(accessToken, {
      spaceId: privateSpace.spaceId,
      title,
      description,
      visibility,
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

function formatDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("ja-JP", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function WorkbookCard({ workbook }: { workbook: Workbook }) {
  const fetcher = useFetcher();
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
        Updated: {formatDate(workbook.updatedAt)}
      </p>

      <div className="flex items-center gap-2">
        <Button size="sm" className="flex-1" disabled>
          <BookOpenIcon data-icon="inline-start" className="size-3.5" />
          <span>Study</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="flex-1"
          render={<Link to={`/workbooks/${workbook.workbookId}`} />}
        >
          <PencilIcon data-icon="inline-start" className="size-3.5" />
          <span>Edit</span>
        </Button>
        <fetcher.Form method="post">
          <input type="hidden" name="intent" value="delete" />
          <input type="hidden" name="workbookId" value={workbook.workbookId} />
          <Button
            variant="destructive"
            size="icon-sm"
            type="submit"
            disabled={isDeleting}
            onClick={(e) => {
              if (!confirm(`Delete "${workbook.title}"?`)) {
                e.preventDefault();
              }
            }}
          >
            <Trash2Icon className="size-3.5" />
            <span className="sr-only">Delete</span>
          </Button>
        </fetcher.Form>
      </div>
    </div>
  );
}

function CreateWorkbookButton() {
  const fetcher = useFetcher();
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
            <span>New Workbook</span>
          </Button>
        }
      />
      <SheetContent side="right">
        <SheetHeader>
          <SheetTitle>Create Workbook</SheetTitle>
          <SheetDescription>Add a new workbook to organize your problems.</SheetDescription>
        </SheetHeader>
        <fetcher.Form key={formKey} method="post" className="flex flex-1 flex-col gap-4 px-4">
          <input type="hidden" name="intent" value="create" />
          <div className="space-y-1.5">
            <label htmlFor="title" className="text-sm font-medium">
              Title <span className="text-destructive">*</span>
            </label>
            <Input id="title" name="title" required maxLength={200} placeholder="e.g. English Vocabulary" />
          </div>
          <div className="space-y-1.5">
            <label htmlFor="description" className="text-sm font-medium">
              Description
            </label>
            <Input id="description" name="description" maxLength={1000} placeholder="Optional description" />
          </div>
          <fieldset className="space-y-1.5">
            <legend className="text-sm font-medium">Visibility</legend>
            <div className="flex gap-4">
              <label className="flex items-center gap-1.5 text-sm">
                <input type="radio" name="visibility" value="private" defaultChecked />
                Private
              </label>
              <label className="flex items-center gap-1.5 text-sm">
                <input type="radio" name="visibility" value="public" />
                Public
              </label>
            </div>
          </fieldset>
          {fetcher.data && !fetcher.data.ok && "error" in fetcher.data && (
            <p className="text-sm text-destructive">{fetcher.data.error}</p>
          )}
          <SheetFooter>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Creating..." : "Create"}
            </Button>
          </SheetFooter>
        </fetcher.Form>
      </SheetContent>
    </Sheet>
  );
}

export default function WorkbooksIndex() {
  const { workbooks } = useLoaderData<typeof loader>();

  return (
    <div>
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">My Workbooks</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Select a workbook to study or manage its problems.
          </p>
        </div>
        <CreateWorkbookButton />
      </div>

      {workbooks.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <BookOpenIcon className="mb-4 size-12 text-muted-foreground/50" />
          <p className="text-lg font-medium text-muted-foreground">No workbooks yet</p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            Create your first workbook to get started.
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
