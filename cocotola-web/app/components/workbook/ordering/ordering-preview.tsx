import {
  chipSegmentIndices,
  type OrderingSegment,
} from "~/components/study/question-types/ordering/answer";
import { slotDisplayTexts } from "~/components/study/question-types/ordering/capitalize";

type PreviewItem = { segmentIndex: number; type: OrderingSegment["type"]; text: string };

function previewItems(segments: readonly OrderingSegment[], lang: string): PreviewItem[] {
  const chips = chipSegmentIndices(segments);
  const texts = slotDisplayTexts(segments, chips, lang);
  return segments.map((segment, segmentIndex) => ({
    segmentIndex,
    type: segment.type,
    text: segment.type === "chip" ? (texts[chips.indexOf(segmentIndex)] ?? "") : segment.text,
  }));
}

export function OrderingPreview({
  segments,
  lang,
  label,
}: {
  segments: readonly OrderingSegment[];
  lang: string;
  label: string;
}) {
  return (
    <ol aria-label={label} className="flex flex-wrap items-center gap-1.5 text-sm">
      {previewItems(segments, lang).map((item) => (
        <li
          key={item.segmentIndex}
          className={item.type === "chip" ? "rounded-md border px-2 py-0.5" : undefined}
        >
          {item.text}
        </li>
      ))}
    </ol>
  );
}
