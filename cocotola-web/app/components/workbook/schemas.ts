import { z, type ZodError } from "zod";

export type QuestionType = "word_fill" | "multiple_choice";

export const choiceSchema = z.object({
  id: z.string(),
  text: z.string().min(1),
  isCorrect: z.boolean(),
});

export type Choice = z.output<typeof choiceSchema>;

export const multipleChoiceFormSchema = z
  .object({
    questionText: z.string().min(1),
    choices: z.array(choiceSchema).min(1),
    explanation: z.string().optional().default(""),
    shuffleChoices: z.boolean(),
  })
  .refine((data) => data.choices.some((c) => c.isCorrect), {
    message: "At least one choice must be marked as correct",
    path: ["choices"],
  });

function formatZodError(error: ZodError): string {
  return error.issues.map((i) => `${i.path.join(".")}: ${i.message}`).join("; ");
}

export const wordFillContentSchema = z.object({
  source: z.object({ text: z.string(), lang: z.string() }).optional(),
  target: z.object({ text: z.string(), lang: z.string() }).optional(),
  explanation: z.string().optional(),
});

export const multipleChoiceContentSchema = z.object({
  questionText: z.string().optional(),
  explanation: z.string().optional(),
  choices: z.array(choiceSchema).optional(),
  shuffleChoices: z.boolean().optional(),
});

export function parseWordFillContent(
  content: string,
): z.output<typeof wordFillContentSchema> | null {
  try {
    return wordFillContentSchema.parse(JSON.parse(content));
  } catch {
    return null;
  }
}

export function parseMultipleChoiceContent(
  content: string,
): z.output<typeof multipleChoiceContentSchema> | null {
  try {
    return multipleChoiceContentSchema.parse(JSON.parse(content));
  } catch {
    return null;
  }
}

export function parseMultipleChoiceFormData(formData: FormData) {
  const questionText = formData.get("questionText");
  if (typeof questionText !== "string" || !questionText.trim()) {
    throw new Response("questionText is required", { status: 400 });
  }

  const choicesJson = formData.get("choices");
  if (typeof choicesJson !== "string") {
    throw new Response("choices is required", { status: 400 });
  }

  let choicesParsed: unknown;
  try {
    choicesParsed = JSON.parse(choicesJson);
  } catch {
    throw new Response("choices must be valid JSON", { status: 400 });
  }

  const result = multipleChoiceFormSchema.safeParse({
    questionText: questionText.trim(),
    choices: choicesParsed,
    explanation:
      typeof formData.get("explanation") === "string" &&
      (formData.get("explanation") as string).trim()
        ? (formData.get("explanation") as string)
        : "",
    shuffleChoices: formData.get("shuffleChoices") === "true",
  });

  if (!result.success) {
    throw new Response(formatZodError(result.error), { status: 400 });
  }

  const parsed = result.data;

  const content = JSON.stringify({
    questionText: parsed.questionText,
    choices: parsed.choices,
    explanation: parsed.explanation,
    displayCount: parsed.choices.length,
    showCorrectCount: false,
    shuffleChoices: parsed.shuffleChoices,
    allowPartialCredit: false,
  });

  const tagsRaw = formData.get("tags");
  const tags =
    typeof tagsRaw === "string"
      ? tagsRaw
          .split(",")
          .map((t) => t.trim())
          .filter(Boolean)
      : [];

  return { content, tags };
}
