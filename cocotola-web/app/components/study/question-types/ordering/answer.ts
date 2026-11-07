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

export function computeRegions(segments: readonly OrderingSegment[]): string[][] {
  return segments.flatMap((s, i) => {
    if (s.type !== "chip" || (i > 0 && segments[i - 1].type === "chip")) return [];
    const end = segments.findIndex((t, j) => j > i && t.type !== "chip");
    return [segments.slice(i, end === -1 ? undefined : end).map((t) => t.text)];
  });
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

function answerRegions(segments: readonly OrderingSegment[], slots: Slots): string[][] | null {
  const placed = slots.filter((chip): chip is number => chip !== null);
  if (placed.length !== slots.length) return null;
  const chips = chipSegmentIndices(segments);
  const answered = segments.map((s, i) =>
    s.type === "chip" ? { ...s, text: segments[placed[chips.indexOf(i)]].text } : s,
  );
  return computeRegions(answered);
}

export function toOrderingDraft(
  segments: readonly OrderingSegment[],
  slots: Slots,
): AnswerDraft<"ordering"> | null {
  const regions = answerRegions(segments, slots);
  if (regions === null) return null;
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
