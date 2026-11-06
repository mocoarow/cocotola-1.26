import type { UniqueIdentifier } from "@dnd-kit/core";
import type { Slots } from "./answer";
import { placeChip, returnChip } from "./placement";

type DropTarget = { kind: "pool" } | { kind: "slot"; slot: number };

export const POOL_DROP_ID = "pool";

const CHIP_ID_PATTERN = /^chip-(\d+)$/;
const SLOT_ID_PATTERN = /^slot-(\d+)$/;

export const chipDragId = (chip: number) => `chip-${chip}`;

export const slotDropId = (slot: number) => `slot-${slot}`;

function parseIndex(id: UniqueIdentifier, pattern: RegExp): number | null {
  if (typeof id !== "string") return null;
  const match = pattern.exec(id);
  return match === null ? null : Number(match[1]);
}

function parseDropTarget(id: UniqueIdentifier, slotCount: number): DropTarget | null {
  if (id === POOL_DROP_ID) return { kind: "pool" };
  const slot = parseIndex(id, SLOT_ID_PATTERN);
  return slot !== null && slot < slotCount ? { kind: "slot", slot } : null;
}

export function applyDrop(
  slots: Slots,
  chips: readonly number[],
  activeId: UniqueIdentifier,
  overId: UniqueIdentifier | null,
): Slots {
  if (overId === null) return slots;
  const chip = parseIndex(activeId, CHIP_ID_PATTERN);
  const target = parseDropTarget(overId, slots.length);
  if (chip === null || target === null || !chips.includes(chip)) return slots;
  if (target.kind === "slot") return placeChip(slots, chip, target.slot);
  const from = slots.indexOf(chip);
  return from === -1 ? slots : returnChip(slots, from);
}
