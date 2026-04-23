import { ArrowLeftIcon, CheckIcon, PencilIcon, PlusIcon, Trash2Icon, XIcon } from "lucide-react";
import { useState } from "react";
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
      return { ok: false, error: "Title is required" };
    }
    if (title.length > 200) {
      return { ok: false, error: "Title must be 200 characters or less" };
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
  const isSubmitting = fetcher.state !== "idle";

  return (
    <fetcher.Form method="post" className="space-y-3">
      <input type="hidden" name="intent" value="updateQuestion" />
      <input type="hidden" name="questionId" value={question.questionId} />
      <input type="hidden" name="orderIndex" value={question.orderIndex} />
      <div className="grid gap-3 sm:grid-cols-2">
        <div className="space-y-1">
          <label htmlFor="edit-sourceLang" className="text-xs font-medium">
            Source Language
          </label>
          <select
            id="edit-sourceLang"
            name="sourceLang"
            defaultValue={parsed.source?.lang ?? "ja"}
            className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
          >
            {LANG_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-sourceText" className="text-xs font-medium">
            Source Text
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
            Target Language
          </label>
          <select
            id="edit-targetLang"
            name="targetLang"
            defaultValue={parsed.target?.lang ?? "en"}
            className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
          >
            {LANG_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <label htmlFor="edit-targetText" className="text-xs font-medium">
            Target Text
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
          Explanation (optional)
        </label>
        <Input id="edit-explanation" name="explanation" defaultValue={parsed.explanation ?? ""} />
      </div>
      <div className="space-y-1">
        <label htmlFor="edit-tags" className="text-xs font-medium">
          Tags (optional)
        </label>
        <Input id="edit-tags" name="tags" defaultValue={question.tags?.join(", ") ?? ""} />
      </div>
      <div className="flex gap-2">
        <Button type="submit" size="sm" disabled={isSubmitting}>
          <CheckIcon data-icon="inline-start" className="size-3.5" />
          <span>{isSubmitting ? "Saving..." : "Save"}</span>
        </Button>
        <Button type="button" size="sm" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </fetcher.Form>
  );
}

function QuestionCard({ question }: { question: Question }) {
  const [editing, setEditing] = useState(false);
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
                <span className="sr-only">Edit question</span>
              </Button>
              <deleteFetcher.Form
                method="post"
                onSubmit={(e) => {
                  if (!window.confirm("Are you sure you want to delete this question?")) {
                    e.preventDefault();
                  }
                }}
              >
                <input type="hidden" name="intent" value="deleteQuestion" />
                <input type="hidden" name="questionId" value={question.questionId} />
                <Button type="submit" size="icon-sm" variant="ghost">
                  <Trash2Icon className="size-3.5 text-destructive" />
                  <span className="sr-only">Delete question</span>
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

const LANG_OPTIONS = [
  { value: "ja", label: "Japanese" },
  { value: "en", label: "English" },
  { value: "it", label: "Italian" },
  { value: "fr", label: "French" },
  { value: "de", label: "German" },
  { value: "es", label: "Spanish" },
  { value: "zh", label: "Chinese" },
  { value: "ko", label: "Korean" },
  { value: "pt", label: "Portuguese" },
];

function AddWordFillForm() {
  const fetcher = useFetcher();
  const isSubmitting = fetcher.state !== "idle";

  return (
    <div className="rounded-lg border bg-card p-5 shadow-sm">
      <h2 className="mb-4 text-base font-semibold">Add Word Fill Question</h2>
      <fetcher.Form method="post" className="space-y-4">
        <input type="hidden" name="intent" value="addWordFill" />

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="sourceLang" className="text-sm font-medium">
              Source Language
            </label>
            <select
              id="sourceLang"
              name="sourceLang"
              defaultValue="ja"
              className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
            >
              {LANG_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label htmlFor="sourceText" className="text-sm font-medium">
              Source Text
            </label>
            <Input id="sourceText" name="sourceText" placeholder="ゴミを捨てる" required />
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <label htmlFor="targetLang" className="text-sm font-medium">
              Target Language
            </label>
            <select
              id="targetLang"
              name="targetLang"
              defaultValue="en"
              className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm"
            >
              {LANG_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label htmlFor="targetText" className="text-sm font-medium">
              Target Text
            </label>
            <Input id="targetText" name="targetText" placeholder="{{throw}} it {{away}}" required />
            <p className="text-xs text-muted-foreground">
              {"Use {{word}} to mark blanks. e.g. {{throw}} it {{away}}"}
            </p>
          </div>
        </div>

        <div className="space-y-2">
          <label htmlFor="explanation" className="text-sm font-medium">
            Explanation (optional)
          </label>
          <Input id="explanation" name="explanation" placeholder="throw away は句動詞です。" />
        </div>

        <div className="space-y-2">
          <label htmlFor="tags" className="text-sm font-medium">
            Tags (optional)
          </label>
          <Input id="tags" name="tags" placeholder="level:beginner,topic:phrasal-verbs" />
          <p className="text-xs text-muted-foreground">Comma-separated, key:value format</p>
        </div>

        <Button type="submit" disabled={isSubmitting}>
          <PlusIcon data-icon="inline-start" className="size-3.5" />
          <span>{isSubmitting ? "Adding..." : "Add Question"}</span>
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
        Back to Workbooks
      </Link>
      {editing ? (
        <fetcher.Form method="post" className="space-y-3">
          <input type="hidden" name="intent" value="updateTitle" />
          <input type="hidden" name="visibility" value={visibility} />
          <div className="space-y-2">
            <label htmlFor="edit-wb-title" className="block text-sm font-medium">
              Title
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
              Description
            </label>
            <Input
              id="edit-wb-description"
              name="description"
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              className="max-w-md"
              placeholder="Workbook description"
            />
          </div>
          <div className="flex items-center gap-2">
            <Button type="submit" size="sm" disabled={isSubmitting}>
              <CheckIcon data-icon="inline-start" className="size-3.5" />
              <span>{isSubmitting ? "Saving..." : "Save"}</span>
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
              Cancel
            </Button>
          </div>
          {fetcher.data && !fetcher.data.ok && "error" in fetcher.data && (
            <p className="text-sm text-destructive">{fetcher.data.error}</p>
          )}
        </fetcher.Form>
      ) : (
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">{title}</h1>
            <Button size="icon-sm" variant="ghost" onClick={() => setEditing(true)}>
              <PencilIcon className="size-4" />
              <span className="sr-only">Edit workbook</span>
            </Button>
          </div>
          {description && (
            <p className="mt-1 text-sm text-muted-foreground">{description}</p>
          )}
        </div>
      )}
    </div>
  );
}

export default function WorkbookDetail() {
  const { workbook, questions } = useLoaderData<typeof loader>();

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
          <p className="text-lg font-medium text-muted-foreground">No questions yet</p>
          <p className="mt-1 text-sm text-muted-foreground/70">
            Use the form above to add your first question.
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
