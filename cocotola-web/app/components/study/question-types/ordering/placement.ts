import type { Slots } from "./answer";

export function emptySlots(count: number): Slots {
  return Array.from({ length: count }, () => null);
}

// Puts the chip into slot `to`. A chip already in another slot swaps with the occupant;
// a chip from the pool sends the occupant back to the pool.
export function placeChip(slots: Slots, chip: number, to: number): Slots {
  const from = slots.indexOf(chip);
  if (from === to) return slots;
  const occupant = slots[to];
  return slots.map((current, i) => {
    if (i === to) return chip;
    if (i === from) return occupant;
    return current;
  });
}

export function placeInFirstEmpty(slots: Slots, chip: number): Slots {
  const to = slots.indexOf(null);
  if (to === -1 || slots.includes(chip)) return slots;
  return placeChip(slots, chip, to);
}

export function returnChip(slots: Slots, from: number): Slots {
  return slots.map((current, i) => (i === from ? null : current));
}
