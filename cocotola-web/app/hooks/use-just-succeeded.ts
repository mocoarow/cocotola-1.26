import { useState } from "react";

type ActionResult = { ok: boolean } | undefined;

export function useJustSucceeded(data: ActionResult): boolean {
  const [handled, setHandled] = useState<ActionResult>(undefined);
  if (data?.ok && data !== handled) {
    setHandled(data);
    return true;
  }
  return false;
}
