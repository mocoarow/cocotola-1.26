import type { OrderingSegment } from "~/components/study/question-types/ordering/answer";
import { MAX_ORDERING_SEGMENTS, MAX_ORDERING_TEXT_BYTES } from "~/lib/study/answer-payload";
import { MIN_ORDERING_CHIPS } from "../schemas";

export type OrderingNotationError =
  | "unmatchedOpen"
  | "unmatchedClose"
  | "emptyChip"
  | "emptyFixed"
  | "tooFewChips"
  | "tooManySegments"
  | "textTooLong"
  | "forbiddenChar"
  | "htmlTag";

export type OrderingNotationResult =
  | { ok: true; segments: OrderingSegment[] }
  | { ok: false; error: OrderingNotationError };

const OPEN = "[[";
const CLOSE = "]]";
const CHIP_SEPARATOR = "/";

// Go's strings.TrimSpace also strips U+0085, which JS \s does not include.
const EDGE_SPACES = /^[\s\u0085]+|[\s\u0085]+$/g;
// Mirror cocotola-question domain/question/content.go htmlTagPattern.
const HTML_TAG = /<[^>]+>/;
const FORBIDDEN_CHARS = /[[\]]/;

const textEncoder = new TextEncoder();

export function trimSpace(text: string): string {
  return text.replace(EDGE_SPACES, "");
}

function chipsOf(text: string): OrderingSegment[] {
  if (trimSpace(text) === "") return [];
  return text.split(CHIP_SEPARATOR).map((chip) => ({ type: "chip", text: trimSpace(chip) }));
}

type FixedPart = { fixed: string; following: string };

function fixedPart(afterOpen: string): FixedPart | OrderingNotationError {
  const end = afterOpen.indexOf(CLOSE);
  if (end === -1) return "unmatchedOpen";
  const following = afterOpen.slice(end + CLOSE.length);
  return following.includes(CLOSE)
    ? "unmatchedClose"
    : { fixed: afterOpen.slice(0, end), following };
}

function splitNotation(notation: string): OrderingNotationResult {
  const [head, ...afterOpens] = notation.split(OPEN);
  if (head.includes(CLOSE)) return { ok: false, error: "unmatchedClose" };
  const parts = afterOpens.map(fixedPart);
  const error = parts.find((part): part is OrderingNotationError => typeof part === "string");
  if (error) return { ok: false, error };
  const fixedParts = parts.filter((part): part is FixedPart => typeof part !== "string");
  const segments = fixedParts.flatMap((part): OrderingSegment[] => [
    { type: "fixed", text: trimSpace(part.fixed) },
    ...chipsOf(part.following),
  ]);
  return { ok: true, segments: [...chipsOf(head), ...segments] };
}

function segmentError(segment: OrderingSegment): OrderingNotationError | null {
  if (segment.text === "") return segment.type === "chip" ? "emptyChip" : "emptyFixed";
  if (HTML_TAG.test(segment.text)) return "htmlTag";
  if (FORBIDDEN_CHARS.test(segment.text)) return "forbiddenChar";
  if (textEncoder.encode(segment.text).length > MAX_ORDERING_TEXT_BYTES) return "textTooLong";
  return null;
}

function segmentsError(segments: readonly OrderingSegment[]): OrderingNotationError | null {
  const error = segments.map(segmentError).find((e) => e !== null);
  if (error) return error;
  if (segments.length > MAX_ORDERING_SEGMENTS) return "tooManySegments";
  if (segments.filter((s) => s.type === "chip").length < MIN_ORDERING_CHIPS) return "tooFewChips";
  return null;
}

export function parseOrderingNotation(notation: string): OrderingNotationResult {
  const result = splitNotation(notation);
  if (!result.ok) return result;
  const error = segmentsError(result.segments);
  return error === null ? result : { ok: false, error };
}

function separatorBefore(segments: readonly OrderingSegment[], index: number): string {
  if (index === 0) return "";
  const bothChips = segments[index].type === "chip" && segments[index - 1].type === "chip";
  return bothChips ? ` ${CHIP_SEPARATOR} ` : " ";
}

export function formatOrderingNotation(segments: readonly OrderingSegment[]): string {
  return segments
    .map((s, i) => {
      const text = s.type === "fixed" ? `${OPEN}${s.text}${CLOSE}` : s.text;
      return separatorBefore(segments, i) + text;
    })
    .join("");
}
