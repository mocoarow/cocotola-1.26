import {
  MouseSensor,
  TouchSensor,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import type { ReactNode } from "react";
import { Button } from "~/components/ui/button";
import { cn } from "~/lib/utils";
import type { BindChip } from "./chip-focus";
import { chipDragId, POOL_DROP_ID, slotDropId } from "./drop";

const MOUSE_DRAG_DISTANCE_PX = 8;
const TOUCH_DRAG_DELAY_MS = 200;
const TOUCH_DRAG_TOLERANCE_PX = 5;

export function useOrderingSensors() {
  return useSensors(
    useSensor(MouseSensor, { activationConstraint: { distance: MOUSE_DRAG_DISTANCE_PX } }),
    useSensor(TouchSensor, {
      activationConstraint: { delay: TOUCH_DRAG_DELAY_MS, tolerance: TOUCH_DRAG_TOLERANCE_PX },
    }),
  );
}

type DraggableChipProps = {
  chip: number;
  bind: BindChip;
  variant: "secondary" | "outline";
  onClick: () => void;
  children: ReactNode;
};

export function DraggableChip({ chip, bind, variant, onClick, children }: DraggableChipProps) {
  const { setNodeRef, listeners, transform, isDragging } = useDraggable({ id: chipDragId(chip) });
  const { ref, ...binding } = bind(chip);
  return (
    <Button
      ref={(el) => {
        ref(el);
        setNodeRef(el);
      }}
      {...listeners}
      {...binding}
      variant={variant}
      onClick={onClick}
      className={cn("touch-manipulation", isDragging && "relative z-10 shadow-md transition-none")}
      style={
        transform ? { transform: `translate3d(${transform.x}px, ${transform.y}px, 0)` } : undefined
      }
    >
      {children}
    </Button>
  );
}

const DROP_OVER_CLASS = "ring-2 ring-primary";

export function DroppableSlot({ slot, children }: { slot: number; children: ReactNode }) {
  const { setNodeRef, isOver } = useDroppable({ id: slotDropId(slot) });
  return (
    <span ref={setNodeRef} className={cn("inline-flex rounded-md", isOver && DROP_OVER_CLASS)}>
      {children}
    </span>
  );
}

export function usePoolDrop() {
  const { setNodeRef, isOver } = useDroppable({ id: POOL_DROP_ID });
  return { ref: setNodeRef, overClass: isOver ? DROP_OVER_CLASS : undefined };
}
