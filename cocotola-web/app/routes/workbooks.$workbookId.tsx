import { ArrowLeftIcon, CheckIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useFetcher, useLoaderData } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import {
  addQuestion,
  deleteQuestion,
  listQuestions,
  type Question,
  updateQuestion,
} from "~/lib/api/question.server";
import { getWorkbook, updateWorkbook } from "~/lib/api/workbook.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const [workbook, questions] = await Promise.all([
    getWorkbook(accessToken, workbookId),
    listQuestions(accessToken, workbookId),
  ]);
  return { workbook, questions };
}

export async function action({ request, params }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const formData = await request.formData();
  const intent = formData.get("intent");

  if (intent === "updateTitle") {
    const title = String(formData.get("title") ?? "").trim();
    if (!title) {
      return { ok: false, errorKey: "workbooks.detail.errors.titleRequired" };
    }
    if (title.length > 200) {
      return { ok: false, errorKey: "workbooks.detail.errors.titleTooLong" };
    }
    const description = String(formData.get("description") ?? "");
    const rawVisibility = String(formData.get("visibility") ?? "private");
    const visibility = rawVisibility === "public" ? "public" : "private";
    await updateWorkbook(accessToken, workbookId, { title, description, visibility });
    return { ok: true };
  }

  if (intent === "updateQuestion") {
    const questionId = formData.get("questionId");
    if (typeof questionId !== "string" || !questionId) {
      throw new Response("questionId is required", { status: 400 });
    }

    const sourceText = formData.get("sourceText");
    const sourceLang = formData.get("sourceLang");
    const targetText = formData.get("targetText");
    const targetLang = formData.get("targetLang");

    if (
      typeof sourceText !== "string" ||
      !sourceText.trim() ||
      typeof sourceLang !== "string" ||
      !sourceLang.trim() ||
      typeof targetText !== "string" ||
      !targetText.trim() ||
      typeof targetLang !== "string" ||
      !targetLang.trim()
    ) {
      throw new Response("sourceText, sourceLang, targetText, and targetLang are required", {
        status: 400,
      });
    }

    const explanation = formData.get("explanation");
    const content = JSON.stringify({
      source: { text: sourceText, lang: sourceLang },
      target: { text: targetText, lang: targetLang },
      ...(typeof explanation === "string" && explanation.trim() ? { explanation } : {}),
    });

    const tagsRaw = formData.get("tags");
    const tags =
      typeof tagsRaw === "string"
        ? tagsRaw
            .split(",")
            .map((t) => t.trim())
            .filter(Boolean)
        : [];

    const orderIndex = Number(formData.get("orderIndex") ?? 0);

    await updateQuestion(accessToken, workbookId, questionId, {
      content,
      tags: tags.length > 0 ? tags : undefined,
      orderIndex,
    });
    return { ok: true };
  }

  if (intent === "deleteQuestion") {
    const questionId = formData.get("questionId");
    if (typeof questionId !== "string" || !questionId) {
      throw new Response("questionId is required", { status: 400 });
    }
    await deleteQuestion(accessToken, workbookId, questionId);
    return { ok: true };
  }

  if (intent === "addWordFill") {
    const sourceText = formData.get("sourceText");
    const sourceLang = formData.get("sourceLang");
    const targetText = formData.get("targetText");
    const targetLang = formData.get("targetLang");

    if (
      typeof sourceText !== "string" ||
      !sourceText.trim() ||
      typeof sourceLang !== "string" ||
      !sourceLang.trim() ||
      typeof targetText !== "string" ||
      !targetText.trim() ||
      typeof targetLang !== "string" ||
      !targetLang.trim()
    ) {
      throw new Response("sourceText, sourceLang, targetText, and targetLang are required", {
        status: 400,
      });
    }

    const explanation = formData.get("explanation");

    const content = JSON.stringify({
      source: { text: sourceText, lang: sourceLang },
      target: { text: targetText, lang: targetLang },
      ...(typeof explanation === "string" && explanation.trim() ? { explanation } : {}),
    });

    const tagsRaw = formData.get("tags");
    const tags =
      typeof tagsRaw === "string"
        ? tagsRaw
            .split(",")
            .map((t) => t.trim())
            .filter(Boolean)
        : [];

    await addQuestion(accessToken, workbookId, {
      questionType: "word_fill",
      content,
      tags: tags.length > 0 ? tags : undefined,
      orderIndex: 0,
    });
  }

  return { ok: true };
}

function parseWordFillContent(content: string): {
  source?: { text: string; lang: string };
  target?: { text: string; lang: string };
  explanation?: string;
} | null {
  try {
    return JSON.parse(content) as {
      source?: { text: string; lang: string };
      target?: { text: string; lang: string };
      explanation?: string;
    };
  } catch {
    return null;
  }
}

const LANG_KEYS = [
  { value: "ja", labelKey: "languages.ja" },
  { value: "en", labelKey: "languages.en" },
  { value: "it", labelKey: "languages.it" },
  { value: "fr", labelKey: "languages.fr" },
  { value: "de", labelKey: "languages.de" },
  { value: "es", labelKey: "languages.es" },
  { value: "zh", labelKey: "languages.zh" },
  { value: "ko", labelKey: "languages.ko" },
  { value: "pt", labelKey: "languages.pt" },
];

function LangSelect({
  id,
  name,
  defaultValue,
}: {
  id: string;
  name: string;
  defaultValue: string;
}) {
  const { t } = useTranslation();
  return (
    <select
      id={id}
      name={name}
      defaultValue={defaultValue}
      className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
    >
      {LANG_KEYS.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {t(opt.labelKey)}
        </option>
      ))}
    </select>
  );
}

function WordFillEditForm({
  question,
  parsed,
  fetcher,
  onCancel,
}: {
  question: Question;
  parsed: {
    source?: { text: string; lang: string };
    target?: { text: string; lang: string };
    explanation?: string;
  };
  fetcher: ReturnType<typeof useFetcher>;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post" className="space-y-3">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <label htmlFor="edit-sourceLang" className="text-xs font-medium">
            {t("workbooks.addQuestion.sourceLang")}
          </label>
          <LangSelect
            id="edit-sourceLang"
            name="sourceLang"
            defaultValue={parsed.source?.lang ?? "ja"}
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-sourceText" className="text-xs font-medium">
            {t("workbooks.addQuestion.sourceText")}
          </label>
          <Input
            id="edit-sourceText"
            name="sourceText"
            defaultValue={parsed.source?.text ?? ""}
            required
          />
        </div>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <label htmlFor="edit-targetLang" className="text-xs font-medium">
            {t("workbooks.addQuestion.targetLang")}
          </label>
          <LangSelect
            id="edit-targetLang"
            name="targetLang"
            defaultValue={parsed.target?.lang ?? "en"}
          />
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-targetText" className="text-xs font-medium">
            {t("workbooks.addQuestion.targetText")}
          </label>
          <Input
            id="edit-targetText"
            name="targetText"
            defaultValue={parsed.target?.text ?? ""}
            required
          />
        </div>
      </div>
      <div className="space-y-1">
        <label htmlFor="edit-explanation" className="text-xs font-medium">
          {t("workbooks.addQuestion.explanation")}
        </label>
        <Input id="edit-explanation" name="explanation" defaultValue={parsed.explanation ?? ""} />
      </div>
      <div className="space-y-1">
        <label htmlFor="edit-tags" className="text-xs font-medium">
          {t("workbooks.addQuestion.tags")}
        </label>
        <Input id="edit-tags" name="tags" defaultValue={question.tags?.join(", ") ?? ""} />
      </div>
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={isSubmitting}>
          <CheckIcon data-icon="inline-start" className="size-3.5" />
          <span>{isSubmitting ? t("common.saving") : t("common.save")}</span>
        </Button>
        <Button type="button" size="sm" variant="outline" onClick={onCancel}>
          {t("common.cancel")}
        </Button>
      </div>
    </fetcher.Form>
  );
}

function QuestionCard({ question }: { question: Question }) {
  const [editing, setEditing] = useState(false);
  const { t } = useTranslation();
  const editFetcher = useFetcher();
  const deleteFetcher = useFetcher();

  if (editFetcher.data?.ok && editing) {
    setEditing(false);
  }

  if (question.questionType === "word_fill") {
    const parsed = parseWordFillContent(question.content);

    return (
      <div className="rounded-lg border bg-card p-4 shadow-sm">
        <div className="mb-2 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="rounded-full bg-blue-100 px-2 py-0.5 text-[11px] font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">
              word_fill
            </span>
            {question.tags?.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] text-gray-600 dark:bg-gray-800 dark:text-gray-400"
              >
                {tag}
              </span>
            ))}
          </div>
          {!editing && (
            <div className="flex items-center gap-1">
              <Button size="icon-sm" variant="ghost" onClick={() => setEditing(true)}>
                <PencilIcon className="size-3.5" />
                <span className="sr-only">{t("workbooks.detail.editQuestion")}</span>
              </Button>
              <deleteFetcher.Form
                method="post"
                onSubmit={(e) => {
                  if (!window.confirm(t("workbooks.detail.deleteQuestionConfirm"))) {
                    e.preventDefault();
                  }
                }}
              >
                <input type="hidden" name="intent" value="deleteQuestion" />
                <input type="hidden" name="questionId" value={question.questionId} />
                <Button type="submit" size="icon-sm" variant="ghost">
                  <Trash2Icon className="size-3.5 text-destructive" />
                  <span className="sr-only">{t("workbooks.detail.deleteQuestion")}</span>
                </Button>
              </deleteFetcher.Form>
            </div>
          )}
        </div>
        {editing && parsed ? (
          <WordFillEditForm
            question={question}
            parsed={parsed}
            fetcher={editFetcher}
            onCancel={() => setEditing(false)}
          />
        ) : parsed ? (
          <div className="space-y-1 text-sm">
            <p>
              <span className="font-medium text-muted-foreground">[{parsed.source?.lang}]</span>{" "}
              {parsed.source?.text}
            </p>
            <p>
              <span className="font-medium text-muted-foreground">[{parsed.target?.lang}]</span>{" "}
              {parsed.target?.text}
            </p>
            {parsed.explanation && (
              <p className="text-xs text-muted-foreground">{parsed.explanation}</p>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{question.content}</p>
        )}
      </div>
    );
  }

  return (
    <div className="rounded-lg border bg-card p-4 shadow-sm">
      <div className="mb-2">
        <span className="rounded-full bg-green-100 px-2 py-0.5 text-[11px] font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">
          {question.questionType}
        </span>
      </div>
      <p className="text-sm text-muted-foreground">{question.content}</p>
    </div>
  );
}

function AddWordFillForm() {
  const fetcher = useFetcher();
  const { t } = useTranslation();
  const isSubmitting = fetcher.state !== "idle";

  return (
    <div className="rounded-lg border bg-card p-5 shadow-sm">
      <h2 className="mb-4 text-base font-semibold">{t("workbooks.addQuestion.title")}</h2>
      <fetcher.Form method="post" className="space-y-4">
        <input type="hidden" name="intent" value="addWordFill" />

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="sourceLang" className="text-sm font-medium">
              {t("workbooks.addQuestion.sourceLang")}
            </label>
            <LangSelect id="sourceLang" name="sourceLang" defaultValue="ja" />
          </div>
          <div className="space-y-2">
            <label htmlFor="sourceText" className="text-sm font-medium">
              {t("workbooks.addQuestion.sourceText")}
            </label>
            <Input id="sourceText" name="sourceText" placeholder="ゴミを捨てる" required />
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="targetLang" className="text-sm font-medium">
              {t("workbooks.addQuestion.targetLang")}
            </label>
            <LangSelect id="targetLang" name="targetLang" defaultValue="en" />
          </div>
          <div className="space-y-2">
            <label htmlFor="targetText" className="text-sm font-medium">
              {t("workbooks.addQuestion.targetText")}
            </label>
            <Input id="targetText" name="targetText" placeholder="{{throw}} it {{away}}" required />
            <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.targetHint")}</p>
          </div>
        </div>

        <div className="space-y-2">
          <label htmlFor="explanation" className="text-sm font-medium">
            {t("workbooks.addQuestion.explanation")}
          </label>
          <Input id="explanation" name="explanation" placeholder="throw away は句動詞です。" />
        </div>

        <div className="space-y-2">
          <label htmlFor="tags" className="text-sm font-medium">
            {t("workbooks.addQuestion.tags")}
          </label>
          <Input id="tags" name="tags" placeholder="level:beginner,topic:phrasal-verbs" />
          <p className="text-xs text-muted-foreground">{t("workbooks.addQuestion.tagsHint")}</p>
        </div>

        <Button type="submit" disabled={isSubmitting}>
          <PlusIcon data-icon="inline-start" className="size-3.5" />
          <span>
            {isSubmitting ? t("workbooks.addQuestion.adding") : t("workbooks.addQuestion.submit")}
          </span>
        </Button>
      </fetcher.Form>
    </div>
  );
}

function WorkbookHeader({
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

export default function WorkbookDetail() {
  const { workbook, questions } = useLoaderData<typeof loader>();
  const { t } = useTranslation();

  return (
    <div>
      <WorkbookHeader
        title={workbook.title}
        description={workbook.description}
        visibility={workbook.visibility}
      />

      <div className="mb-6">
        <AddWordFillForm />
      </div>

      {questions.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg border border-dashed py-16">
          <p className="text-lg font-medium text-muted-foreground">
            {t("workbooks.detail.emptyTitle")}
          </p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            {t("workbooks.detail.emptyDescription")}
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {questions.map((question) => (
            <QuestionCard key={question.questionId} question={question} />
          ))}
        </div>
      )}
    </div>
  );
}
