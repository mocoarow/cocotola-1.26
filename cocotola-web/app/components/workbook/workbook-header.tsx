import { ArrowLeftIcon, CheckIcon, PencilIcon } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useFetcher } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";

export function WorkbookHeader({
  title,
  description,
  visibility,
}: {
  title: string;
  description: string;
  visibility: string;
}) {
  const fetcher = useFetcher();
  const { t } = useTranslation();
  const [editing, setEditing] = useState(false);
  const [editTitle, setEditTitle] = useState(title);
  const [editDescription, setEditDescription] = useState(description);
  const isSubmitting = fetcher.state !== "idle";

  if (fetcher.data?.ok && editing) {
    setEditing(false);
  }

  return (
    <div className="mb-6">
      <Link
        to="/workbooks"
        className="mb-2 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeftIcon className="size-3.5" />
        {t("workbooks.detail.backToWorkbooks")}
      </Link>
      {editing ? (
        <fetcher.Form method="post" className="space-y-3">
          <input type="hidden" name="intent" value="updateTitle" />
          <input type="hidden" name="visibility" value={visibility} />
          <div className="space-y-2">
            <label htmlFor="edit-wb-title" className="block text-sm font-medium">
              {t("workbooks.detail.titleLabel")}
            </label>
            <Input
              id="edit-wb-title"
              name="title"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              className="max-w-md text-lg font-bold"
              maxLength={200}
              required
              autoFocus
            />
          </div>
          <div className="space-y-2">
            <label htmlFor="edit-wb-description" className="block text-sm font-medium">
              {t("workbooks.detail.descriptionLabel")}
            </label>
            <Input
              id="edit-wb-description"
              name="description"
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              className="max-w-md"
              placeholder={t("workbooks.detail.descriptionPlaceholder")}
            />
          </div>
          <div className="flex items-center gap-2">
            <Button type="submit" size="sm" disabled={isSubmitting}>
              <CheckIcon data-icon="inline-start" className="size-3.5" />
              <span>{isSubmitting ? t("common.saving") : t("common.save")}</span>
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => {
                setEditing(false);
                setEditTitle(title);
                setEditDescription(description);
              }}
            >
              {t("common.cancel")}
            </Button>
          </div>
          {fetcher.data && !fetcher.data.ok && "errorKey" in fetcher.data && (
            <p className="text-sm text-destructive">{t(fetcher.data.errorKey)}</p>
          )}
        </fetcher.Form>
      ) : (
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{title}</h1>
            <Button size="icon-sm" variant="ghost" onClick={() => setEditing(true)}>
              <PencilIcon className="size-4" />
              <span className="sr-only">{t("workbooks.detail.editWorkbook")}</span>
            </Button>
          </div>
          {description && <p className="mt-1 text-sm text-muted-foreground">{description}</p>}
        </div>
      )}
    </div>
  );
}
