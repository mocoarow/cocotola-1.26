import { chipSegmentIndices, type OrderingSegment, type Slots } from "./answer";

const SENTENCE_END = /[.?!]["'”’»)」』）]*$/;

function upperFirst(text: string, lang: string): string {
  const [first = ""] = text;
  const rest = text.slice(first.length);
  try {
    return first.toLocaleUpperCase(lang) + rest;
  } catch (error) {
    if (error instanceof RangeError) return first.toUpperCase() + rest;
    throw error;
  }
}

// Text visible right before the segment, or null when it is an empty slot.
function precedingText(
  segments: readonly OrderingSegment[],
  slots: Slots,
  segmentIndex: number,
  slot: number,
): string | null {
  const previous = segments[segmentIndex - 1];
  if (previous.type === "fixed") return previous.text;
  const chip = slots[slot - 1];
  return chip === null ? null : segments[chip].text;
}

function isSentenceStart(
  segments: readonly OrderingSegment[],
  slots: Slots,
  segmentIndex: number,
  slot: number,
): boolean {
  if (segmentIndex === 0) return true;
  const text = precedingText(segments, slots, segmentIndex, slot);
  return text !== null && SENTENCE_END.test(text);
}

export function slotDisplayTexts(
  segments: readonly OrderingSegment[],
  slots: Slots,
  lang: string,
): (string | null)[] {
  return chipSegmentIndices(segments).map((segmentIndex, slot) => {
    const chip = slots[slot];
    if (chip === null) return null;
    const text = segments[chip].text;
    return isSentenceStart(segments, slots, segmentIndex, slot) ? upperFirst(text, lang) : text;
  });
}
