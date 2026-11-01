import type { QuestionModule } from "~/components/study/question-shell/types";
import {
  extractBlanks,
  isCorrectAnswer,
  parseRenderableWordFill,
  toWordFillDraft,
  type WordFillContent,
} from "./answer";
import { WordFillBody } from "./body";

export const wordFillModule: QuestionModule<"word_fill", WordFillContent> = {
  parse: parseRenderableWordFill,
  initialDraft: (content) => {
    const { answers } = extractBlanks(content.target.text);
    return toWordFillDraft(answers.length > 0 && answers.every((a) => isCorrectAnswer("", a)));
  },
  revealLabelKey: "workbooks.study.showAnswer",
  explanation: (content, phase) =>
    (phase === "revealed" ? content.explanation2 || content.explanation : content.explanation1) ||
    undefined,
  Body: WordFillBody,
};
