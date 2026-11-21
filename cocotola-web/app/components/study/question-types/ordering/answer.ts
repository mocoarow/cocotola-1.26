import type { AnswerDraft } from "~/components/study/question-shell/types";
import type { parseOrderingContent } from "~/components/workbook/schemas";

export type OrderingContent = NonNullable<ReturnType<typeof parseOrderingContent>>;
export type OrderingSegment = OrderingContent["target"]["segments"][number];

// A chip is identified by its index in `segments`. Slot k belongs to the k-th chip segment
// and holds the chip placed there, or null while empty.
export type Slots = readonly (number | null)[];

export function chipSegmentIndices(segments: readonly OrderingSegment[]): number[] {
  return segments.flatMap((s, i) => (s.type === "chip" ? [i] : []));
}

function regionSegmentIndices(segments: readonly OrderingSegment[]): number[][] {
  return segments.flatMap((s, i) => {
    if (s.type !== "chip" || (i > 0 && segments[i - 1].type === "chip")) return [];
    const end = segments.findIndex((t, j) => j > i && t.type !== "chip");
    const length = (end === -1 ? segments.length : end) - i;
    return [Array.from({ length }, (_, k) => i + k)];
  });
}

export function computeRegions(segments: readonly OrderingSegment[]): string[][] {
  return regionSegmentIndices(segments).map((region) => region.map((i) => segments[i].text));
}

export function evaluateRegions(
  segments: readonly OrderingSegment[],
  answer: readonly (readonly string[])[],
): boolean {
  const want = computeRegions(segments);
  return (
    want.length === answer.length &&
    want.every(
      (region, i) =>
        region.length === answer[i].length && region.every((text, j) => text === answer[i][j]),
    )
  );
}

// Empty slots are left out, so an answer with an empty slot never matches the correct regions.
function answerRegions(segments: readonly OrderingSegment[], slots: Slots): string[][] {
  const chips = chipSegmentIndices(segments);
  return regionSegmentIndices(segments).map((region) =>
    region.flatMap((i) => {
      const chip = slots[chips.indexOf(i)];
      return chip === null ? [] : [segments[chip].text];
    }),
  );
}

export function toOrderingDraft(
  segments: readonly OrderingSegment[],
  slots: Slots,
): AnswerDraft<"ordering"> {
  const regions = answerRegions(segments, slots);
  return {
    correct: evaluateRegions(segments, regions),
    payload: { type: "ordering", regions },
  };
}

export function slotCorrectness(segments: readonly OrderingSegment[], slots: Slots): boolean[] {
  return chipSegmentIndices(segments).map((segmentIndex, slot) => {
    const chip = slots[slot];
    return chip !== null && segments[chip].text === segments[segmentIndex].text;
  });
}
