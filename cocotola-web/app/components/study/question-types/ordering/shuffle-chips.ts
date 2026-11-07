import { shuffle } from "~/components/study/question-types/shuffle";
import { chipSegmentIndices, type OrderingSegment } from "./answer";

const MAX_SHUFFLE_ATTEMPTS = 10;

function sameTexts(
  segments: readonly OrderingSegment[],
  a: readonly number[],
  b: readonly number[],
): boolean {
  return a.every((chip, i) => segments[chip].text === segments[b[i]].text);
}

// Returns the chip indices in an order whose texts differ from the correct order,
// unless every chip has the same text.
export function shuffleChips(segments: readonly OrderingSegment[]): number[] {
  const correct = chipSegmentIndices(segments);
  const allSame = correct.every((chip) => segments[chip].text === segments[correct[0]].text);
  if (allSame) return shuffle(correct);
  for (let attempt = 0; attempt < MAX_SHUFFLE_ATTEMPTS; attempt++) {
    const order = shuffle(correct);
    if (!sameTexts(segments, order, correct)) return order;
  }
  // A rotation of a non-constant sequence never equals the sequence itself.
  return [...correct.slice(1), correct[0]];
}
