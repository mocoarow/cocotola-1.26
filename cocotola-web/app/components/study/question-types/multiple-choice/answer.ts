import type { AnswerDraft } from "~/components/study/question-shell/types";
import type { parseMultipleChoiceContent } from "~/components/workbook/schemas";

export type MultipleChoiceContent = NonNullable<ReturnType<typeof parseMultipleChoiceContent>>;

function setsEqual(a: ReadonlySet<string>, b: ReadonlySet<string>): boolean {
  if (a.size !== b.size) return false;
  for (const v of a) {
    if (!b.has(v)) return false;
  }
  return true;
}

export function toMultipleChoiceDraft(
  selectedIds: ReadonlySet<string>,
  correctIds: ReadonlySet<string>,
): AnswerDraft<"multiple_choice"> | null {
  if (selectedIds.size === 0) return null;
  return {
    correct: setsEqual(selectedIds, correctIds),
    payload: { type: "multiple_choice", selectedChoiceIds: Array.from(selectedIds) },
  };
}
