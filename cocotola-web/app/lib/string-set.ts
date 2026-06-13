/**
 * Returns true when the two arrays describe the same string set, ignoring
 * duplicates and order. Used by the study resume orchestrator to decide
 * whether the URL's excludeIds already cover the locally stored answered
 * IDs; a length-only comparison would misfire on duplicates (e.g. ["x","y"]
 * vs ["x","x"]).
 */
export function sameStringSet(a: readonly string[], b: readonly string[]): boolean {
  const sa = new Set(a);
  const sb = new Set(b);
  if (sa.size !== sb.size) return false;
  for (const v of sa) {
    if (!sb.has(v)) return false;
  }
  return true;
}
