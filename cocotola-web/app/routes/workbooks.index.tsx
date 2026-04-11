import { BookOpenIcon, PencilIcon, Trash2Icon } from "lucide-react";
import { useFetcher, useLoaderData } from "react-router";
import { Button } from "~/components/ui/button";
import { findPrivateSpace } from "~/lib/api/space.server";
import { deleteWorkbook, listWorkbooks, type Workbook } from "~/lib/api/workbook.server";
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
        <Button variant="outline" size="sm" className="flex-1" disabled>
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

export default function WorkbooksIndex() {
  const { workbooks } = useLoaderData<typeof loader>();

  return (
    <div>
      <div className="mb-6">
        <h1 className="text-2xl font-bold">My Workbooks</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Select a workbook to study or manage its problems.
        </p>
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
