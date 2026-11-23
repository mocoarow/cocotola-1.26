import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  deleteStudyHistory,
  type GetStudyQuestionsResponse,
  getStudyQuestions,
  getStudySummary,
  listStudyRecords,
  type RecordAnswerResponse,
  recordAnswer,
  type StudyQuestion,
  type StudyRecord,
  type StudySummary,
} from "./study.server";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";

const WORD_FILL_PAYLOAD = { type: "word_fill", correct: true } as const;

const LOCALE = { localDateKey: "2026-11-23", timezone: "America/Los_Angeles" };

function buildStudyQuestion(overrides: Partial<StudyQuestion> = {}): StudyQuestion {
  return {
    questionId: "q-42",
    questionType: "word_fill",
    content: JSON.stringify({
      source: { text: "ゴミを捨てる", lang: "ja" },
      target: { text: "{{throw}} it {{away}}", lang: "en" },
    }),
    tags: ["level:beginner", "topic:phrasal-verbs"],
    orderIndex: 3,
    audio: {
      source: { url: "https://cdn.example.com/audio/q-42-source.mp3", durationSec: 1.25 },
      target: { url: "https://cdn.example.com/audio/q-42-target.mp3", durationSec: 2.5 },
    },
    ...overrides,
  };
}

function buildStudyQuestionsResponse(): GetStudyQuestionsResponse {
  return {
    questions: [buildStudyQuestion()],
    totalDue: 17,
    newCount: 5,
    reviewCount: 12,
  };
}

function buildStudySummary(): StudySummary {
  return {
    newCount: 5,
    reviewCount: 12,
    totalDue: 17,
    reviewRatioNumerator: 7,
    reviewRatioDenominator: 11,
  };
}

function buildRecordAnswerResponse(): RecordAnswerResponse {
  return {
    nextDueAt: "2026-11-24T09:30:00Z",
    consecutiveCorrect: 3,
    totalCorrect: 7,
    totalIncorrect: 2,
  };
}

function buildStudyRecord(): StudyRecord {
  return {
    workbookId: "wb-1",
    questionId: "q-42",
    consecutiveCorrect: 2,
    lastAnsweredAt: "2026-11-20T10:00:00Z",
    nextDueAt: "2026-11-22T10:00:00Z",
    totalCorrect: 5,
    totalIncorrect: 1,
  };
}

beforeEach(() => {
  vi.stubEnv("QUESTION_BASE_URL", QUESTION_BASE_URL);
  vi.spyOn(console, "info").mockImplementation(() => {});
  vi.spyOn(console, "error").mockImplementation(() => {});
});

afterEach(() => {
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
  fetchMock.mockReset();
});

describe("getStudyQuestions", () => {
  it("should return study questions on success", async () => {
    // given
    const response = buildStudyQuestionsResponse();
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when
    const result = await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(result).toEqual(response);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=20`,
      { method: "GET", headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should clamp limit to valid range", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", 200);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=100`,
      expect.any(Object),
    );
  });

  it("should clamp limit minimum to 1", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", -5);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=1`,
      expect.any(Object),
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb/1", 10);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/study?limit=10`,
      expect.any(Object),
    );
  });

  it("should pass practice=true query when requested", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", 10, true);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=10&practice=true`,
      expect.any(Object),
    );
  });

  it("should append repeated excludeIds query params", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", 10, false, ["q-1", "q-2"]);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=10&excludeIds=q-1&excludeIds=q-2`,
      expect.any(Object),
    );
  });

  it("should skip empty-string entries in excludeIds", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudyQuestionsResponse()));

    // when
    await getStudyQuestions(newClient(), "wb-1", 10, false, ["", "q-1"]);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study?limit=10&excludeIds=q-1`,
      expect.any(Object),
    );
  });

  it("should accept a question without tags", async () => {
    // given
    const { tags: _, ...withoutTags } = buildStudyQuestion();
    fetchMock.mockResolvedValue(
      jsonResponse({ ...buildStudyQuestionsResponse(), questions: [withoutTags] }),
    );

    // when
    const result = await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(result.questions).toEqual([withoutTags]);
  });

  it("should accept a question without audio", async () => {
    // given
    const { audio: _, ...withoutAudio } = buildStudyQuestion();
    fetchMock.mockResolvedValue(
      jsonResponse({ ...buildStudyQuestionsResponse(), questions: [withoutAudio] }),
    );

    // when
    const result = await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(result.questions).toEqual([withoutAudio]);
  });

  it("should accept audio that lacks the source ref", async () => {
    // given
    const question = buildStudyQuestion({
      audio: { target: { url: "https://cdn.example.com/audio/q-42-target.mp3", durationSec: 2.5 } },
    });
    fetchMock.mockResolvedValue(
      jsonResponse({ ...buildStudyQuestionsResponse(), questions: [question] }),
    );

    // when
    const result = await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(result.questions).toEqual([question]);
  });

  it("should accept audio refs whose urls are empty", async () => {
    // given
    const question = buildStudyQuestion({
      audio: { source: { url: "", durationSec: 0 }, target: { url: "", durationSec: 0 } },
    });
    fetchMock.mockResolvedValue(
      jsonResponse({ ...buildStudyQuestionsResponse(), questions: [question] }),
    );

    // when
    const result = await getStudyQuestions(newClient(), "wb-1", 20);

    // then
    expect(result.questions).toEqual([question]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(getStudyQuestions(newClient(), "wb-1", 20)).rejects.toMatchObject({
      status: 500,
    });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(getStudyQuestions(newClient(), "wb-1", 20));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a question in the response lacks content", async () => {
    // given
    const { content: _, ...withoutContent } = buildStudyQuestion();
    fetchMock.mockResolvedValue(
      jsonResponse({ ...buildStudyQuestionsResponse(), questions: [withoutContent] }),
    );

    // when / then
    await expect(getStudyQuestions(newClient(), "wb-1", 20)).rejects.toMatchObject({
      status: 502,
    });
  });

  it("should throw 502 Response when the response has a non-numeric totalDue", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ...buildStudyQuestionsResponse(), totalDue: "17" }));

    // when / then
    await expect(getStudyQuestions(newClient(), "wb-1", 20)).rejects.toMatchObject({
      status: 502,
    });
  });
});

describe("getStudySummary", () => {
  it("should return summary on success", async () => {
    // given
    const summary = buildStudySummary();
    fetchMock.mockResolvedValue(jsonResponse(summary));

    // when
    const result = await getStudySummary(newClient(), "wb-1");

    // then
    expect(result).toEqual(summary);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudySummary()));

    // when
    await getStudySummary(newClient(), "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study/summary`,
      { method: "GET", headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should pass practice=true query when requested", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudySummary()));

    // when
    await getStudySummary(newClient(), "wb-1", true);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study/summary?practice=true`,
      expect.any(Object),
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildStudySummary()));

    // when
    await getStudySummary(newClient(), "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/study/summary`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(getStudySummary(newClient(), "wb-1")).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(getStudySummary(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks reviewRatioNumerator", async () => {
    // given
    const { reviewRatioNumerator: _, ...withoutNumerator } = buildStudySummary();
    fetchMock.mockResolvedValue(jsonResponse(withoutNumerator));

    // when / then
    await expect(getStudySummary(newClient(), "wb-1")).rejects.toMatchObject({ status: 502 });
  });
});

describe("recordAnswer", () => {
  it("should return the recorded answer on success", async () => {
    // given
    const response = buildRecordAnswerResponse();
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when
    const result = await recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD);

    // then
    expect(result).toEqual(response);
  });

  it("should send POST request with correct=true body when payload is word_fill", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study/q-1/answer`,
      {
        method: "POST",
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ correct: true }),
      },
    );
  });

  it("should send correct=false when word_fill answer is incorrect", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", { type: "word_fill", correct: false });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ body: JSON.stringify({ correct: false }) }),
    );
  });

  it("should send selectedChoiceIds body when payload is multiple_choice", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", {
      type: "multiple_choice",
      selectedChoiceIds: ["c1", "c2"],
    });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ body: JSON.stringify({ selectedChoiceIds: ["c1", "c2"] }) }),
    );
  });

  it("should send empty selectedChoiceIds when nothing selected", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", {
      type: "multiple_choice",
      selectedChoiceIds: [],
    });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ body: JSON.stringify({ selectedChoiceIds: [] }) }),
    );
  });

  it("should send regions body when payload is ordering", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", {
      type: "ordering",
      regions: [["I", "am"], ["you"]],
    });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ body: JSON.stringify({ regions: [["I", "am"], ["you"]] }) }),
    );
  });

  it("should send the locale headers when locale is given", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD, LOCALE);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
          "X-Local-Date": "2026-11-23",
          "X-Local-Timezone": "America/Los_Angeles",
        },
      }),
    );
  });

  it.each([
    ["the local date is empty", { ...LOCALE, localDateKey: "" }],
    ["the timezone is empty", { ...LOCALE, timezone: "" }],
  ])("should not send the locale headers when %s", async (_, locale) => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD, locale);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        headers: { Authorization: "Bearer test-token", "Content-Type": "application/json" },
      }),
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildRecordAnswerResponse()));

    // when
    await recordAnswer(newClient(), "wb/1", "q/2", WORD_FILL_PAYLOAD);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/study/q%2F2/answer`,
      expect.any(Object),
    );
  });

  it.each([400, 409, 500])("should keep the status %i in the Response", async (status) => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status }));

    // when / then
    await expect(recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD)).rejects.toMatchObject(
      { status },
    );
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks nextDueAt", async () => {
    // given
    const { nextDueAt: _, ...withoutNextDueAt } = buildRecordAnswerResponse();
    fetchMock.mockResolvedValue(jsonResponse(withoutNextDueAt));

    // when / then
    await expect(recordAnswer(newClient(), "wb-1", "q-1", WORD_FILL_PAYLOAD)).rejects.toMatchObject(
      { status: 502 },
    );
  });
});

describe("deleteStudyHistory", () => {
  it("should send DELETE request with correct URL and headers", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteStudyHistory(newClient(), "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study`, {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteStudyHistory(newClient(), "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/study`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(deleteStudyHistory(newClient(), "wb-1")).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(deleteStudyHistory(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});

describe("listStudyRecords", () => {
  it("should return records on success", async () => {
    // given
    const response = { records: [buildStudyRecord()] };
    fetchMock.mockResolvedValue(jsonResponse(response));

    // when
    const result = await listStudyRecords(newClient(), "wb-1");

    // then
    expect(result).toEqual(response);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ records: [] }));

    // when
    await listStudyRecords(newClient(), "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/study/records`,
      { method: "GET", headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ records: [] }));

    // when
    await listStudyRecords(newClient(), "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/study/records`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(listStudyRecords(newClient(), "wb-1")).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(listStudyRecords(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a record in the response lacks questionId", async () => {
    // given
    const { questionId: _, ...withoutQuestionId } = buildStudyRecord();
    fetchMock.mockResolvedValue(jsonResponse({ records: [withoutQuestionId] }));

    // when / then
    await expect(listStudyRecords(newClient(), "wb-1")).rejects.toMatchObject({ status: 502 });
  });
});
