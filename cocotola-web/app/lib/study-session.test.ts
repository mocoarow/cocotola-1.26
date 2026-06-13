import { describe, expect, it } from "vitest";
import {
  appendAnsweredId,
  clampExcludeIds,
  clearStudySession,
  isStudySessionStale,
  loadStudySession,
  MAX_EXCLUDE_ID_LENGTH,
  MAX_EXCLUDE_IDS_COUNT,
  nextLocal3am,
  type StudySessionScope,
  saveStudySession,
  studyRemountKey,
  studySessionStorageKey,
} from "./study-session";

function fakeStorage(): Storage {
  const map = new Map<string, string>();
  return {
    get length() {
      return map.size;
    },
    clear: () => map.clear(),
    getItem: (k: string) => (map.has(k) ? (map.get(k) ?? null) : null),
    key: (i: number) => Array.from(map.keys())[i] ?? null,
    removeItem: (k: string) => {
      map.delete(k);
    },
    setItem: (k: string, v: string) => {
      map.set(k, v);
    },
  };
}

const scopeA: StudySessionScope = { userId: "user-a", workbookId: "wb-1", practice: false };
const scopeB: StudySessionScope = { userId: "user-b", workbookId: "wb-1", practice: false };

describe("studySessionStorageKey", () => {
  it("should produce distinct keys for different users on the same workbook", () => {
    // when
    const a = studySessionStorageKey(scopeA);
    const b = studySessionStorageKey(scopeB);

    // then
    expect(a).not.toBe(b);
  });

  it("should produce distinct keys for practice and normal modes", () => {
    // given/when
    const normal = studySessionStorageKey(scopeA);
    const practice = studySessionStorageKey({ ...scopeA, practice: true });

    // then
    expect(normal).not.toBe(practice);
  });

  it("should not collide when userId or workbookId contains the ':' delimiter", () => {
    // Regression: an unencoded key built with `${userId}:${workbookId}` would
    // map both inputs to "cocotola:study:a:b:c", letting the two users'
    // saves overwrite each other (an availability bug — payload checks block
    // cross-user reads but not writes).

    // given/when
    const split = studySessionStorageKey({ userId: "a:b", workbookId: "c", practice: false });
    const joined = studySessionStorageKey({ userId: "a", workbookId: "b:c", practice: false });

    // then
    expect(split).not.toBe(joined);
  });

  it("should not collide when userId starts with the 'practice:' mode prefix", () => {
    // Without segment encoding, a userId of "practice:foo" with practice=false
    // would build the same key as userId="foo" with practice=true.

    // given/when
    const userIdLooksLikePrefix = studySessionStorageKey({
      userId: "practice:foo",
      workbookId: "wb-1",
      practice: false,
    });
    const actualPractice = studySessionStorageKey({
      userId: "foo",
      workbookId: "wb-1",
      practice: true,
    });

    // then
    expect(userIdLooksLikePrefix).not.toBe(actualPractice);
  });
});

describe("nextLocal3am", () => {
  it("should return the same-day 03:00 when started before 03:00 local", () => {
    // given: 2026-01-15 01:30 local
    const started = new Date(2026, 0, 15, 1, 30, 0).getTime();

    // when
    const boundary = nextLocal3am(started);

    // then
    const b = new Date(boundary);
    expect(b.getHours()).toBe(3);
    expect(b.getDate()).toBe(15);
  });

  it("should return the next-day 03:00 when started after 03:00 local", () => {
    // given: 2026-01-15 10:00 local
    const started = new Date(2026, 0, 15, 10, 0, 0).getTime();

    // when
    const boundary = nextLocal3am(started);

    // then
    const b = new Date(boundary);
    expect(b.getHours()).toBe(3);
    expect(b.getDate()).toBe(16);
  });
});

describe("isStudySessionStale", () => {
  it("should be false when now is before the next 03:00 boundary", () => {
    // given: started 22:00, check at 02:59 next day
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    const now = new Date(2026, 0, 16, 2, 59, 0).getTime();

    // when/then
    expect(isStudySessionStale(started, now)).toBe(false);
  });

  it("should be true when now is past the next 03:00 boundary", () => {
    // given: started 22:00, check at 03:01 next day
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    const now = new Date(2026, 0, 16, 3, 1, 0).getTime();

    // when/then
    expect(isStudySessionStale(started, now)).toBe(true);
  });
});

describe("loadStudySession", () => {
  it("should return null when no session is stored", () => {
    // given
    const storage = fakeStorage();

    // when
    const state = loadStudySession(scopeA, 0, storage);

    // then
    expect(state).toBeNull();
  });

  it("should return the stored state when fresh", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: ["q-1"] }, storage);

    // when: check at 02:59 (before next 03:00)
    const now = new Date(2026, 0, 16, 2, 59, 0).getTime();
    const state = loadStudySession(scopeA, now, storage);

    // then
    expect(state).not.toBeNull();
    expect(state?.answeredIds).toEqual(["q-1"]);
  });

  it("should return null and remove storage when stale", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: ["q-1"] }, storage);

    // when: check at 03:01 next day (past boundary)
    const now = new Date(2026, 0, 16, 3, 1, 0).getTime();
    const state = loadStudySession(scopeA, now, storage);

    // then
    expect(state).toBeNull();
    expect(storage.getItem(studySessionStorageKey(scopeA))).toBeNull();
  });

  it("should return null when stored JSON is malformed", () => {
    // given
    const storage = fakeStorage();
    storage.setItem(studySessionStorageKey(scopeA), "{not json");

    // when
    const state = loadStudySession(scopeA, 0, storage);

    // then
    expect(state).toBeNull();
  });

  it("should not surface user A's session when user B asks (key isolation)", () => {
    // given: only user A has a stored session.
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: ["q-1"] }, storage);

    // when: user B queries on the same workbook
    const now = new Date(2026, 0, 15, 22, 30, 0).getTime();
    const state = loadStudySession(scopeB, now, storage);

    // then: B sees nothing — no cross-user resume leakage.
    expect(state).toBeNull();
  });

  it("should refuse a payload whose userId disagrees with the asking scope", () => {
    // given: a stored record under user A's scope key, but with the userId
    // field forged to point at user B. parseState accepts it; loadStudySession
    // must catch the mismatch and drop it.
    const storage = fakeStorage();
    storage.setItem(
      studySessionStorageKey(scopeA),
      JSON.stringify({
        userId: scopeB.userId,
        workbookId: scopeA.workbookId,
        practice: false,
        startedAtMs: Date.now(),
        answeredIds: ["q-1"],
      }),
    );

    // when
    const state = loadStudySession(scopeA, Date.now(), storage);

    // then
    expect(state).toBeNull();
    expect(storage.getItem(studySessionStorageKey(scopeA))).toBeNull();
  });

  it("should return null when the scope userId is empty", () => {
    // given: anonymous / pre-auth call sites must not surface anything.
    const storage = fakeStorage();
    saveStudySession({ ...scopeA, startedAtMs: Date.now(), answeredIds: ["q-1"] }, storage);

    // when
    const state = loadStudySession({ ...scopeA, userId: "" }, Date.now(), storage);

    // then
    expect(state).toBeNull();
  });
});

describe("saveStudySession", () => {
  it("should be a no-op when the state userId is empty", () => {
    // given
    const storage = fakeStorage();

    // when
    saveStudySession({ ...scopeA, userId: "", startedAtMs: Date.now(), answeredIds: [] }, storage);

    // then
    expect(storage.getItem(studySessionStorageKey({ ...scopeA, userId: "" }))).toBeNull();
  });
});

describe("appendAnsweredId", () => {
  it("should append the questionId when the session is fresh", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: [] }, storage);

    // when
    const now = new Date(2026, 0, 15, 22, 5, 0).getTime();
    appendAnsweredId(scopeA, "q-1", now, storage);

    // then
    const state = loadStudySession(scopeA, now, storage);
    expect(state?.answeredIds).toEqual(["q-1"]);
  });

  it("should not duplicate when the same questionId is appended twice", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: ["q-1"] }, storage);

    // when
    const now = new Date(2026, 0, 15, 22, 5, 0).getTime();
    appendAnsweredId(scopeA, "q-1", now, storage);

    // then
    const state = loadStudySession(scopeA, now, storage);
    expect(state?.answeredIds).toEqual(["q-1"]);
  });

  it("should be a no-op when the session is stale", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: [] }, storage);

    // when: now is past the boundary
    const now = new Date(2026, 0, 16, 3, 1, 0).getTime();
    appendAnsweredId(scopeA, "q-1", now, storage);

    // then: the stale state was cleared and nothing new was written
    expect(storage.getItem(studySessionStorageKey(scopeA))).toBeNull();
  });
});

describe("studyRemountKey", () => {
  it("should produce the same key regardless of excludeIds order", () => {
    // given/when
    const a = studyRemountKey("user-a", false, ["q-2", "q-1"]);
    const b = studyRemountKey("user-a", false, ["q-1", "q-2"]);

    // then: set-equivalent inputs map to the same key so we don't trigger a
    // spurious remount when the URL serializes the IDs in a different order.
    expect(a).toBe(b);
  });

  it('should not confuse ["a,b"] with ["a", "b"]', () => {
    // Regression: a join(",") implementation would produce the same key for
    // both inputs even though they are semantically distinct sets.

    // given/when
    const combined = studyRemountKey("user-a", false, ["a,b"]);
    const split = studyRemountKey("user-a", false, ["a", "b"]);

    // then
    expect(combined).not.toBe(split);
  });

  it("should distinguish between users", () => {
    // given/when
    const a = studyRemountKey("user-a", false, ["q-1"]);
    const b = studyRemountKey("user-b", false, ["q-1"]);

    // then
    expect(a).not.toBe(b);
  });

  it("should distinguish between practice and normal modes", () => {
    // given/when
    const normal = studyRemountKey("user-a", false, ["q-1"]);
    const practice = studyRemountKey("user-a", true, ["q-1"]);

    // then
    expect(normal).not.toBe(practice);
  });

  it("should distinguish a null userId from a user literally named like the sentinel", () => {
    // given/when: even if a userId looks like a sentinel string we might
    // pick for null (e.g. "anon", "null"), JSON encoding the tuple keeps the
    // two cases separate — null serializes as `null`, a string as `"..."`.
    const nullUser = studyRemountKey(null, false, []);
    const stringUser = studyRemountKey("null", false, []);

    // then
    expect(nullUser).not.toBe(stringUser);
  });

  it("should not be confused by a userId containing the legacy ':' delimiter", () => {
    // Regression: a string-concatenated key like `${userId}:${mode}:...`
    // would collide between userId="a" + practice=false and userId="a:normal"
    // + practice=true once we ran the colon-split parser in reverse.

    // given/when
    const plain = studyRemountKey("a", true, []);
    const withColon = studyRemountKey("a:normal", false, []);

    // then
    expect(plain).not.toBe(withColon);
  });
});

describe("clampExcludeIds", () => {
  it("should drop empty entries", () => {
    // given
    const input = ["", "q-1", "", "q-2"];

    // when
    const result = clampExcludeIds(input);

    // then
    expect(result).toEqual(["q-1", "q-2"]);
  });

  it("should drop entries longer than the per-element cap", () => {
    // given
    const tooLong = "a".repeat(MAX_EXCLUDE_ID_LENGTH + 1);
    const ok = "a".repeat(MAX_EXCLUDE_ID_LENGTH);

    // when
    const result = clampExcludeIds([tooLong, ok, "q-1"]);

    // then
    expect(result).toEqual([ok, "q-1"]);
  });

  it("should truncate the slice to the count cap", () => {
    // given
    const input = Array.from({ length: MAX_EXCLUDE_IDS_COUNT + 5 }, (_, i) => `q-${i}`);

    // when
    const result = clampExcludeIds(input);

    // then
    expect(result).toHaveLength(MAX_EXCLUDE_IDS_COUNT);
    expect(result[0]).toBe("q-0");
    expect(result[MAX_EXCLUDE_IDS_COUNT - 1]).toBe(`q-${MAX_EXCLUDE_IDS_COUNT - 1}`);
  });

  it("should return an empty array when given an empty array", () => {
    // when/then
    expect(clampExcludeIds([])).toEqual([]);
  });
});

describe("clearStudySession", () => {
  it("should remove the stored session", () => {
    // given
    const storage = fakeStorage();
    const started = new Date(2026, 0, 15, 22, 0, 0).getTime();
    saveStudySession({ ...scopeA, startedAtMs: started, answeredIds: ["q-1"] }, storage);

    // when
    clearStudySession(scopeA, storage);

    // then
    expect(storage.getItem(studySessionStorageKey(scopeA))).toBeNull();
  });
});
