import { CheckIcon, XIcon } from "lucide-react";
import { useId } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { chipSegmentIndices, type OrderingSegment, type Slots, slotCorrectness } from "./answer";
import { slotDisplayTexts } from "./capitalize";

export type RegisterChip = (chip: number) => (el: HTMLButtonElement | null) => void;

const CHIP_CLASS = "inline-flex h-9 items-center gap-1 rounded-md border px-3 text-base";

type Token =
  | { kind: "fixed"; segmentIndex: number; text: string }
  | { kind: "slot"; segmentIndex: number; slot: number };

function layoutTokens(segments: readonly OrderingSegment[]): Token[] {
  const chips = chipSegmentIndices(segments);
  return segments.map((s, segmentIndex) =>
    s.type === "fixed"
      ? { kind: "fixed", segmentIndex, text: s.text }
      : { kind: "slot", segmentIndex, slot: chips.indexOf(segmentIndex) },
  );
}

function EmptySlot({ slot }: { slot: number }) {
  const { t } = useTranslation();
  return (
    <span className="inline-block h-9 w-16 rounded-md border-2 border-dashed border-muted-foreground/40">
      <span className="sr-only">{t("workbooks.study.blankInput", { number: slot + 1 })}</span>
    </span>
  );
}

function RevealedSlot({ text, correct }: { text: string; correct: boolean }) {
  const { t } = useTranslation();
  const Icon = correct ? CheckIcon : XIcon;
  const colorClass = correct
    ? "border-green-500 bg-green-50 text-green-900 dark:bg-green-950/30 dark:text-green-300"
    : "border-red-500 bg-red-50 text-red-900 dark:bg-red-950/30 dark:text-red-300";
  return (
    <span className={`${CHIP_CLASS} ${colorClass}`}>
      {text}
      <Icon aria-hidden="true" className="size-4" />
      <span className="sr-only">
        {correct
          ? t("workbooks.study.ordering.slotCorrect")
          : t("workbooks.study.ordering.slotIncorrect")}
      </span>
    </span>
  );
}

type AnswerLineProps = {
  segments: readonly OrderingSegment[];
  slots: Slots;
  lang: string;
  revealed: boolean;
  registerSlot: RegisterChip;
  onSlotTap: (slot: number) => void;
};

export function AnswerLine({
  segments,
  slots,
  lang,
  revealed,
  registerSlot,
  onSlotTap,
}: AnswerLineProps) {
  const { t } = useTranslation();
  const texts = slotDisplayTexts(segments, slots, lang);
  const correctness = slotCorrectness(segments, slots);
  return (
    <fieldset
      aria-label={t("workbooks.study.ordering.answer")}
      className="flex flex-wrap items-center gap-2 text-lg"
    >
      {layoutTokens(segments).map((token) => {
        if (token.kind === "fixed") return <span key={token.segmentIndex}>{token.text}</span>;
        const chip = slots[token.slot];
        const text = texts[token.slot];
        if (chip === null || text === null) {
          return <EmptySlot key={token.segmentIndex} slot={token.slot} />;
        }
        if (revealed) {
          return (
            <RevealedSlot key={token.segmentIndex} text={text} correct={correctness[token.slot]} />
          );
        }
        return (
          <Button
            key={token.segmentIndex}
            ref={registerSlot(chip)}
            variant="outline"
            onClick={() => onSlotTap(token.slot)}
          >
            {text}
          </Button>
        );
      })}
    </fieldset>
  );
}

export function CorrectOrder({
  segments,
  lang,
}: {
  segments: readonly OrderingSegment[];
  lang: string;
}) {
  const { t } = useTranslation();
  const headingId = useId();
  const texts = slotDisplayTexts(segments, chipSegmentIndices(segments), lang);
  return (
    <div className="space-y-2">
      <p id={headingId} className="text-sm font-medium text-muted-foreground">
        {t("workbooks.study.ordering.correctOrder")}
      </p>
      <ol aria-labelledby={headingId} className="flex flex-wrap items-center gap-2 text-lg">
        {layoutTokens(segments).map((token) =>
          token.kind === "fixed" ? (
            <li key={token.segmentIndex}>{token.text}</li>
          ) : (
            <li key={token.segmentIndex} className={`${CHIP_CLASS} border-green-500`}>
              {texts[token.slot]}
            </li>
          ),
        )}
      </ol>
    </div>
  );
}
