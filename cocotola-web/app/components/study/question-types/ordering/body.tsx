import { type Announcements, DndContext, type DragEndEvent } from "@dnd-kit/core";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import type { BodyProps } from "~/components/study/question-shell/types";
import { cn } from "~/lib/utils";
import { type OrderingContent, type OrderingSegment, type Slots, toOrderingDraft } from "./answer";
import { AnswerLine, CorrectOrder } from "./answer-line";
import { type BindChip, useChipFocus } from "./chip-focus";
import { DraggableChip, useOrderingSensors, usePoolDrop } from "./drag";
import { applyDrop } from "./drop";
import { chipAreas } from "./keyboard";
import { emptySlots, placeInFirstEmpty, returnChip } from "./placement";
import { shuffleChips } from "./shuffle-chips";

const NO_ANNOUNCEMENTS: Announcements = {
  onDragStart: () => undefined,
  onDragMove: () => undefined,
  onDragOver: () => undefined,
  onDragEnd: () => undefined,
  onDragCancel: () => undefined,
};

type ChipPoolProps = {
  segments: readonly OrderingSegment[];
  chips: readonly number[];
  helpId: string;
  bind: BindChip;
  onTap: (chip: number) => void;
};

function ChipPool({ segments, chips, helpId, bind, onTap }: ChipPoolProps) {
  const { t } = useTranslation();
  const drop = usePoolDrop();
  return (
    <div
      ref={drop.ref}
      role="toolbar"
      aria-label={t("workbooks.study.ordering.chips")}
      aria-describedby={helpId}
      className={cn(
        "flex min-h-12 flex-wrap gap-2 rounded-lg border border-dashed p-2",
        drop.overClass,
      )}
    >
      {chips.map((chip) => (
        <DraggableChip
          key={chip}
          chip={chip}
          bind={bind}
          variant="secondary"
          onClick={() => onTap(chip)}
        >
          {segments[chip].text}
        </DraggableChip>
      ))}
    </div>
  );
}

function KeyboardHelp({ id }: { id: string }) {
  const { t } = useTranslation();
  return (
    <p id={id} data-testid="ordering-keyboard-help" className="text-sm text-muted-foreground">
      {t("workbooks.study.ordering.keyboardHelp")}
    </p>
  );
}

function useOrderingSlots(
  segments: readonly OrderingSegment[],
  onDraftChange: BodyProps<"ordering", OrderingContent>["onDraftChange"],
) {
  const [order] = useState(() => shuffleChips(segments));
  const [slots, setSlots] = useState<Slots>(() => emptySlots(order.length));
  const poolChips = order.filter((chip) => !slots.includes(chip));
  const chipFocus = useChipFocus(chipAreas(poolChips, slots));

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

  function handleDragEnd({ active, over }: DragEndEvent) {
    const next = applyDrop(slots, order, active.id, over?.id ?? null);
    if (next !== slots) update(next);
  }

  return { slots, poolChips, chipFocus, handlePoolTap, handleSlotTap, handleDragEnd };
}

export function OrderingBody({
  content,
  phase,
  onDraftChange,
}: BodyProps<"ordering", OrderingContent>) {
  const { segments, lang } = content.target;
  const { slots, poolChips, chipFocus, handlePoolTap, handleSlotTap, handleDragEnd } =
    useOrderingSlots(segments, onDraftChange);
  const sensors = useOrderingSensors();
  const helpId = useId();
  const revealed = phase === "revealed";

  return (
    <DndContext
      sensors={sensors}
      accessibility={{ announcements: NO_ANNOUNCEMENTS }}
      onDragEnd={revealed ? undefined : handleDragEnd}
    >
      <div className="rounded-lg bg-muted/50 p-4">
        <p className="text-lg font-medium">{content.source.text}</p>
      </div>
      <AnswerLine
        segments={segments}
        slots={slots}
        lang={lang}
        revealed={revealed}
        helpId={helpId}
        bindSlot={chipFocus.bindSlot}
        onSlotTap={handleSlotTap}
      />
      {revealed ? (
        <CorrectOrder segments={segments} lang={lang} />
      ) : (
        <>
          <ChipPool
            segments={segments}
            chips={poolChips}
            helpId={helpId}
            bind={chipFocus.bindPool}
            onTap={handlePoolTap}
          />
          <KeyboardHelp id={helpId} />
        </>
      )}
    </DndContext>
  );
}
