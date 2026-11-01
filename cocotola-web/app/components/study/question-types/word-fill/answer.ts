import type { AnswerDraft } from "~/components/study/question-shell/types";
import { parseWordFillContent } from "~/components/workbook/schemas";

type ParsedWordFill = NonNullable<ReturnType<typeof parseWordFillContent>>;

export type WordFillContent = ParsedWordFill & {
  target: NonNullable<ParsedWordFill["target"]>;
};

export function parseRenderableWordFill(raw: string): WordFillContent | null {
  const parsed = parseWordFillContent(raw);
  if (!parsed?.target?.text) return null;
  return { ...parsed, target: parsed.target };
}

export function extractBlanks(text: string): { segments: string[]; answers: string[] } {
  const segments = text.split(/\{\{([^}]+)\}\}/g);
  const answers = segments.filter((_, i) => i % 2 === 1);
  return { segments, answers };
}

export function isCorrectAnswer(value: string, answer: string): boolean {
  return value.trim().toLowerCase() === answer.trim().toLowerCase();
}

// Returns the first blank after `from` (exclusive, wrapping) that is not yet correct, or -1.
export function findNextUnlocked(from: number, correct: readonly boolean[]): number {
  for (let step = 1; step <= correct.length; step++) {
    const candidate = (from + step) % correct.length;
    if (!correct[candidate]) return candidate;
  }
  return -1;
}

export function toWordFillDraft(correct: boolean): AnswerDraft<"word_fill"> {
  return { correct, payload: { type: "word_fill", correct } };
}
