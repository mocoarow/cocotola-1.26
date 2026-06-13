/**
 * Browser-side resume support for the study screen.
 *
 * The server is stateless: GetStudyQuestions returns a fresh random selection on
 * every call. To make F5/back-button feel like "continue from where I left off",
 * the browser tracks the in-progress session in localStorage and replays the
 * already-answered IDs to the server via `excludeIds`. The set is discarded
 * when the user crosses the next local-time 03:00 boundary so the "study day"
 * still resets nightly.
 *
 * Every key and persisted value is scoped to (userId, workbookId, practice) so
 * that signing out user A and signing in user B on the same browser never
 * surfaces A's answeredIds in B's URL or queue. The userId is *also* checked
 * inside the parsed payload so a stale entry written under a key that was
 * later reused (e.g. by a userId collision attack) is still rejected.
 *
 * No network access here — only pure storage + time math, so the same helpers
 * can be unit-tested without a DOM mock.
 */

/**
 * Mirrors the server's GetStudyQuestionsInput caps so URL-supplied IDs that
 * would be rejected with a 400 can be dropped before the API call is issued.
 * Keep these in sync with cocotola-question/service/study/study.go
 * (MaxExcludeIDsCount / MaxExcludeIDLength).
 */
export const MAX_EXCLUDE_IDS_COUNT = 1000;
export const MAX_EXCLUDE_ID_LENGTH = 100;

/**
 * Normalises a list of excludeIds taken from URL search params before sending
 * them to the API. Drops empty entries and any ID longer than the server cap,
 * then truncates the survivors to the count cap. Defensive only — the server
 * validates again as the trust boundary.
 */
export function clampExcludeIds(raw: readonly string[]): string[] {
  const filtered: string[] = [];
  for (const id of raw) {
    if (id === "") continue;
    if (id.length > MAX_EXCLUDE_ID_LENGTH) continue;
    filtered.push(id);
    if (filtered.length === MAX_EXCLUDE_IDS_COUNT) break;
  }
  return filtered;
}

export type StudySessionScope = {
  userId: string;
  workbookId: string;
  practice: boolean;
};

export type StudySessionState = StudySessionScope & {
  startedAtMs: number;
  answeredIds: string[];
};

const STORAGE_PREFIX = "cocotola:study:";

export function studySessionStorageKey(scope: StudySessionScope): string {
  // Encode each segment so a ':' (or any reserved char) inside userId or
  // workbookId cannot collide with the delimiter. Without this, userId="a:b"
  // + workbookId="c" and userId="a" + workbookId="b:c" would share a key and
  // the two users' save() calls would overwrite each other. The payload
  // checks in loadStudySession block cross-user reads, but availability
  // (lost progress until the next 03:00 reset) was still at risk.
  const mode = scope.practice ? "practice:" : "";
  const userId = encodeURIComponent(scope.userId);
  const workbookId = encodeURIComponent(scope.workbookId);
  return `${STORAGE_PREFIX}${mode}${userId}:${workbookId}`;
}

type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;

function getStorage(): StorageLike | null {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

/**
 * Returns the next 03:00 boundary (in milliseconds) that falls strictly after
 * `startedAtMs`, evaluated in the browser's local timezone. Sessions are
 * discarded once `Date.now()` reaches this boundary.
 *
 * Exported for unit testing — the host can stub a fixed `Date` to assert the
 * boundary lands on the right local day across DST and timezone changes.
 */
export function nextLocal3am(startedAtMs: number): number {
  const d = new Date(startedAtMs);
  const candidate = new Date(d.getFullYear(), d.getMonth(), d.getDate(), 3, 0, 0, 0);
  if (candidate.getTime() <= startedAtMs) {
    candidate.setDate(candidate.getDate() + 1);
  }
  return candidate.getTime();
}

export function isStudySessionStale(startedAtMs: number, nowMs: number): boolean {
  return nowMs >= nextLocal3am(startedAtMs);
}

function parseState(raw: string | null): StudySessionState | null {
  if (raw === null) return null;
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof parsed !== "object" || parsed === null) return null;
  const o = parsed as Record<string, unknown>;
  if (typeof o.userId !== "string") return null;
  if (typeof o.workbookId !== "string") return null;
  if (typeof o.practice !== "boolean") return null;
  if (typeof o.startedAtMs !== "number" || !Number.isFinite(o.startedAtMs)) return null;
  if (!Array.isArray(o.answeredIds) || !o.answeredIds.every((v) => typeof v === "string")) {
    return null;
  }
  return {
    userId: o.userId,
    workbookId: o.workbookId,
    practice: o.practice,
    startedAtMs: o.startedAtMs,
    answeredIds: [...(o.answeredIds as string[])],
  };
}

export function loadStudySession(
  scope: StudySessionScope,
  nowMs: number,
  storage: StorageLike | null = getStorage(),
): StudySessionState | null {
  if (storage === null) return null;
  if (scope.userId === "") return null;
  const key = studySessionStorageKey(scope);
  const state = parseState(storage.getItem(key));
  if (state === null) return null;
  // Defense in depth: even with userId-scoped keys, refuse to surface a record
  // whose payload disagrees with the asking scope. Drops cross-user leakage if
  // the key namespace ever collides (e.g. future migration bug).
  if (
    state.userId !== scope.userId ||
    state.workbookId !== scope.workbookId ||
    state.practice !== scope.practice
  ) {
    storage.removeItem(key);
    return null;
  }
  if (isStudySessionStale(state.startedAtMs, nowMs)) {
    storage.removeItem(key);
    return null;
  }
  return state;
}

export function saveStudySession(
  state: StudySessionState,
  storage: StorageLike | null = getStorage(),
): void {
  if (storage === null) return;
  if (state.userId === "") return;
  const key = studySessionStorageKey(state);
  storage.setItem(key, JSON.stringify(state));
}

export function clearStudySession(
  scope: StudySessionScope,
  storage: StorageLike | null = getStorage(),
): void {
  if (storage === null) return;
  if (scope.userId === "") return;
  storage.removeItem(studySessionStorageKey(scope));
}

/**
 * Builds the React `key` for the study screen's StudySession component.
 *
 * Folds in the session-defining inputs so a change forces a remount, which is
 * required to re-seed the in-memory queue from a fresh `questions` prop:
 *
 *   - userId  → log-out/log-in on a shared browser resets the queue.
 *   - practice flag → toggling SRS-off practice mode resets the queue.
 *   - excludeIds → a resume-driven navigate finishes by re-mounting with the
 *     post-exclusion question list (otherwise the StudySession's lazy useState
 *     keeps the pre-exclusion queue and re-shows answered cards).
 *
 * The whole tuple is JSON-encoded — never string-concatenated with a
 * separator — so no part of the input can collide with another by containing
 * a delimiter. The excludeIds slice is sorted so set-equivalent inputs (the
 * URL serializer does not guarantee order) map to the same key, matching the
 * set semantics used elsewhere (sameStringSet, server-side filterExcluded).
 */
export function studyRemountKey(
  userId: string | null,
  practice: boolean,
  excludeIds: readonly string[],
): string {
  return JSON.stringify([userId, practice, [...excludeIds].sort()]);
}

export function appendAnsweredId(
  scope: StudySessionScope,
  questionId: string,
  nowMs: number,
  storage: StorageLike | null = getStorage(),
): void {
  if (storage === null) return;
  if (scope.userId === "") return;
  const state = loadStudySession(scope, nowMs, storage);
  if (state === null) return;
  if (state.answeredIds.includes(questionId)) return;
  const next: StudySessionState = {
    ...state,
    answeredIds: [...state.answeredIds, questionId],
  };
  saveStudySession(next, storage);
}
