import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { BodyProps } from "~/components/study/question-shell/types";
import { Button } from "~/components/ui/button";
import { type OrderingContent, type OrderingSegment, type Slots, toOrderingDraft } from "./answer";
import { AnswerLine, CorrectOrder, type RegisterChip } from "./answer-line";
import { emptySlots, placeInFirstEmpty, returnChip } from "./placement";
import { shuffleChips } from "./shuffle-chips";

type ChipArea = "pool" | "slot";

// Keeps keyboard focus on a chip button after the chip moves between the pool and a slot.
function useChipFocus() {
  const buttons = useRef(new Map<string, HTMLButtonElement>());
  const [target, setTarget] = useState<{ key: string } | null>(null);

  useEffect(() => {
    if (target !== null) buttons.current.get(target.key)?.focus();
  }, [target]);

  const register =
    (area: ChipArea): RegisterChip =>
    (chip) =>
    (el) => {
      if (el) buttons.current.set(`${area}-${chip}`, el);
      else buttons.current.delete(`${area}-${chip}`);
    };

  return {
    registerPool: register("pool"),
    registerSlot: register("slot"),
    focus: (area: ChipArea, chip: number) => setTarget({ key: `${area}-${chip}` }),
  };
}

type ChipPoolProps = {
  segments: readonly OrderingSegment[];
  chips: readonly number[];
  register: RegisterChip;
  onTap: (chip: number) => void;
};

function ChipPool({ segments, chips, register, onTap }: ChipPoolProps) {
  const { t } = useTranslation();
  return (
    <fieldset
      aria-label={t("workbooks.study.ordering.chips")}
      className="flex min-h-12 flex-wrap gap-2 rounded-lg border border-dashed p-2"
    >
      {chips.map((chip) => (
        <Button key={chip} ref={register(chip)} variant="secondary" onClick={() => onTap(chip)}>
          {segments[chip].text}
        </Button>
      ))}
    </fieldset>
  );
}

export function OrderingBody({
  content,
  phase,
  onDraftChange,
}: BodyProps<"ordering", OrderingContent>) {
  const { segments, lang } = content.target;
  const [order] = useState(() => shuffleChips(segments));
  const [slots, setSlots] = useState<Slots>(() => emptySlots(order.length));
  const chipFocus = useChipFocus();
  const revealed = phase === "revealed";
  const poolChips = order.filter((chip) => !slots.includes(chip));

  function update(next: Slots) {
    setSlots(next);
    onDraftChange(toOrderingDraft(segments, next));
  }

  function handlePoolTap(chip: number) {
    const next = placeInFirstEmpty(slots, chip);
    if (next === slots) return;
    update(next);
    const remaining = order.filter((c) => !next.includes(c));
    if (remaining.length === 0) {
      chipFocus.focus("slot", chip);
      return;
    }
    chipFocus.focus("pool", remaining[Math.min(poolChips.indexOf(chip), remaining.length - 1)]);
  }

  function handleSlotTap(slot: number) {
    const chip = slots[slot];
    if (chip === null) return;
    update(returnChip(slots, slot));
    chipFocus.focus("pool", chip);
  }

  return (
    <>
      <div className="rounded-lg bg-muted/50 p-4">
        <p className="text-lg font-medium">{content.source.text}</p>
      </div>
      <AnswerLine
        segments={segments}
        slots={slots}
        lang={lang}
        revealed={revealed}
        registerSlot={chipFocus.registerSlot}
        onSlotTap={handleSlotTap}
      />
      {revealed ? (
        <CorrectOrder segments={segments} lang={lang} />
      ) : (
        <ChipPool
          segments={segments}
          chips={poolChips}
          register={chipFocus.registerPool}
          onTap={handlePoolTap}
        />
      )}
    </>
  );
}
