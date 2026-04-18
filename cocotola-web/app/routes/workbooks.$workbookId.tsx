import { ArrowLeftIcon, PlusIcon } from "lucide-react";
import { Link, useFetcher, useLoaderData } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { addQuestion, listQuestions, type Question } from "~/lib/api/question.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks.$workbookId";

export async function loader({ request, params }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const questions = await listQuestions(accessToken, workbookId);
  return { questions };
}

export async function action({ request, params }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const { workbookId } = params;
  const formData = await request.formData();
  const intent = formData.get("intent");

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

function QuestionCard({ question }: { question: Question }) {
  if (question.questionType === "word_fill") {
    const parsed = parseWordFillContent(question.content);
    return (
      <div className="rounded-lg border bg-card p-4 shadow-sm">
        <div className="mb-2 flex items-center gap-2">
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
        {parsed ? (
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

export default function WorkbookDetail() {
  const { questions } = useLoaderData<typeof loader>();

  return (
    <div>
      <div className="mb-6">
        <Link
          to="/workbooks"
          className="mb-2 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeftIcon className="size-3.5" />
          Back to Workbooks
        </Link>
        <h1 className="text-2xl font-bold">Questions</h1>
        <p className="mt-1 text-sm text-muted-foreground">Manage questions in this workbook.</p>
      </div>

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
