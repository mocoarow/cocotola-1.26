import { type Announcements, DndContext, type DragEndEvent } from "@dnd-kit/core";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { BodyProps } from "~/components/study/question-shell/types";
import { cn } from "~/lib/utils";
import { type OrderingContent, type OrderingSegment, type Slots, toOrderingDraft } from "./answer";
import { AnswerLine, CorrectOrder } from "./answer-line";
import { DraggableChip, type RegisterChip, useOrderingSensors, usePoolDrop } from "./drag";
import { applyDrop } from "./drop";
import { emptySlots, placeInFirstEmpty, returnChip } from "./placement";
import { shuffleChips } from "./shuffle-chips";

type ChipArea = "pool" | "slot";

const NO_ANNOUNCEMENTS: Announcements = {
  onDragStart: () => undefined,
  onDragMove: () => undefined,
  onDragOver: () => undefined,
  onDragEnd: () => undefined,
  onDragCancel: () => undefined,
};

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
  const drop = usePoolDrop();
  return (
    <fieldset
      ref={drop.ref}
      aria-label={t("workbooks.study.ordering.chips")}
      className={cn(
        "flex min-h-12 flex-wrap gap-2 rounded-lg border border-dashed p-2",
        drop.overClass,
      )}
    >
      {chips.map((chip) => (
        <DraggableChip
          key={chip}
          chip={chip}
          register={register}
          variant="secondary"
          onClick={() => onTap(chip)}
        >
          {segments[chip].text}
        </DraggableChip>
      ))}
    </fieldset>
  );
}

function useOrderingSlots(
  segments: readonly OrderingSegment[],
  onDraftChange: BodyProps<"ordering", OrderingContent>["onDraftChange"],
) {
  const [order] = useState(() => shuffleChips(segments));
  const [slots, setSlots] = useState<Slots>(() => emptySlots(order.length));
  const chipFocus = useChipFocus();
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
    </DndContext>
  );
}
