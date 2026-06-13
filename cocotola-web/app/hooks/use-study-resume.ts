import { useEffect, useLayoutEffect, useRef, useState } from "react";
import type { NavigateFunction } from "react-router";

import { sameStringSet } from "~/lib/string-set";
import { loadStudySession, type StudySessionScope, saveStudySession } from "~/lib/study-session";

/**
 * Run as `useLayoutEffect` on the client (so the resume decision lands before
 * paint and the pre-resume answer card is never shown) and as `useEffect` on
 * the server (so we don't trip React's "useLayoutEffect does nothing on the
 * server" warning during SSR). Both hooks are no-ops if there is no DOM, so
 * the substitution is safe.
 */
const useIsomorphicLayoutEffect = typeof window === "undefined" ? useEffect : useLayoutEffect;

export type StudyResumeStatus = "ready" | "redirecting";

export type UseStudyResumeArgs = {
  userId: string | null;
  workbookId: string;
  practice: boolean;
  loaderExcludeIds: string[];
  navigate: NavigateFunction;
};

/**
 * Mounts the per-tab study session: reads the locally tracked answered IDs,
 * decides whether to keep, drop, or replay them, and surfaces a "redirecting"
 * status while the loader rerun is in flight so the caller can hide the
 * answer card and prevent double-submits.
 *
 * The server is stateless — we replay the answered IDs through the URL so the
 * loader's GetStudyQuestions call excludes them. Skipped entirely when the
 * userId is null: callers are expected to normalize empty strings to null at
 * the boundary so this hook only needs the one check. The storage key is
 * scoped per user to prevent A→B login leakage on a shared browser. The
 * study route's server loader already requireAuth's, so a null userId here
 * means the layout loader hasn't materialized in this render path — better
 * to study without resume than to risk surfacing another user's answered IDs.
 */
export function useStudyResume({
  userId,
  workbookId,
  practice,
  loaderExcludeIds,
  navigate,
}: UseStudyResumeArgs): StudyResumeStatus {
  const [status, setStatus] = useState<StudyResumeStatus>("ready");
  const didOrchestrateRef = useRef(false);

  useIsomorphicLayoutEffect(() => {
    if (didOrchestrateRef.current) return;
    didOrchestrateRef.current = true;
    if (userId === null) return;

    const scope: StudySessionScope = { userId, workbookId, practice };
    const now = Date.now();
    const stored = loadStudySession(scope, now);

    if (stored === null) {
      // No session or stale (already cleared by loadStudySession). Seed a
      // fresh one. If the URL still carries stale excludeIds from a previous
      // (now-expired) day, drop them so the fresh session sees the full pool.
      saveStudySession({ ...scope, startedAtMs: now, answeredIds: [] });
      if (loaderExcludeIds.length > 0) {
        const url = new URL(window.location.href);
        url.searchParams.delete("excludeIds");
        setStatus("redirecting");
        navigate(`${url.pathname}${url.search}`, { replace: true });
      }
      return;
    }

    // Live session. If the URL already reflects the stored answered set we're
    // done; otherwise navigate to sync them so the loader's next call honors
    // the resume state.
    if (stored.answeredIds.length === 0) return;
    if (sameStringSet(stored.answeredIds, loaderExcludeIds)) return;

    const url = new URL(window.location.href);
    url.searchParams.delete("excludeIds");
    for (const id of stored.answeredIds) {
      url.searchParams.append("excludeIds", id);
    }
    setStatus("redirecting");
    navigate(`${url.pathname}${url.search}`, { replace: true });
  }, [userId, workbookId, practice, loaderExcludeIds, navigate]);

  return status;
}
