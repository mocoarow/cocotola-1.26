import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  getStudyQuestions,
  getStudySummary,
  recordAnswerForMultipleChoice,
  recordAnswerForWordFill,
} from "./study.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("getStudyQuestions", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return study questions on success", async () => {
    // given
    const response = {
      questions: [{ questionId: "q-1", questionType: "word_fill", content: "{}", orderIndex: 0 }],
      totalDue: 1,
      newCount: 1,
      reviewCount: 0,
    };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(response),
    });

    // when
    const result = await getStudyQuestions("test-token", "wb-1", 20);

    // then
    expect(result).toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study?limit=20",
      { headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should clamp limit to valid range", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions: [], totalDue: 0, newCount: 0, reviewCount: 0 }),
    });

    // when
    await getStudyQuestions("token", "wb-1", 200);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study?limit=100",
      expect.any(Object),
    );
  });

  it("should clamp limit minimum to 1", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions: [], totalDue: 0, newCount: 0, reviewCount: 0 }),
    });

    // when
    await getStudyQuestions("token", "wb-1", -5);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study?limit=1",
      expect.any(Object),
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions: [], totalDue: 0, newCount: 0, reviewCount: 0 }),
    });

    // when
    await getStudyQuestions("token", "wb/1", 10);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/study?limit=10",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(getStudyQuestions("token", "wb-1", 20)).rejects.toBeInstanceOf(Response);
  });
});

describe("getStudySummary", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return summary on success", async () => {
    // given
    const response = {
      newCount: 5,
      reviewCount: 12,
      totalDue: 17,
      reviewRatioNumerator: 9,
      reviewRatioDenominator: 10,
    };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(response),
    });

    // when
    const result = await getStudySummary("test-token", "wb-1");

    // then
    expect(result).toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study/summary",
      { headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should pass practice=true query when requested", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve({
          newCount: 0,
          reviewCount: 0,
          totalDue: 0,
          reviewRatioNumerator: 9,
          reviewRatioDenominator: 10,
        }),
    });

    // when
    await getStudySummary("token", "wb-1", true);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study/summary?practice=true",
      expect.any(Object),
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () =>
        Promise.resolve({
          newCount: 0,
          reviewCount: 0,
          totalDue: 0,
          reviewRatioNumerator: 9,
          reviewRatioDenominator: 10,
        }),
    });

    // when
    await getStudySummary("token", "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/study/summary",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(getStudySummary("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});

describe("recordAnswerForWordFill", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request with correct=true body", async () => {
    // given
    const response = {
      nextDueAt: "2026-04-26T00:00:00Z",
      consecutiveCorrect: 1,
      totalCorrect: 1,
      totalIncorrect: 0,
    };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(response),
    });

    // when
    const result = await recordAnswerForWordFill("test-token", "wb-1", "q-1", true);

    // then
    expect(result).toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study/q-1/answer",
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

  it("should send correct=false for incorrect answer", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    await recordAnswerForWordFill("token", "wb-1", "q-1", false);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        body: JSON.stringify({ correct: false }),
      }),
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    await recordAnswerForWordFill("token", "wb/1", "q/2", true);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/study/q%2F2/answer",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 400 });

    // when / then
    await expect(recordAnswerForWordFill("token", "wb-1", "q-1", true)).rejects.toBeInstanceOf(
      Response,
    );
  });
});

describe("recordAnswerForMultipleChoice", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request with selectedChoiceIds body", async () => {
    // given
    const response = {
      nextDueAt: "2026-04-26T00:00:00Z",
      consecutiveCorrect: 1,
      totalCorrect: 1,
      totalIncorrect: 0,
    };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(response),
    });

    // when
    const result = await recordAnswerForMultipleChoice("test-token", "wb-1", "q-1", ["c1", "c2"]);

    // then
    expect(result).toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study/q-1/answer",
      {
        method: "POST",
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ selectedChoiceIds: ["c1", "c2"] }),
      },
    );
  });

  it("should send empty selectedChoiceIds when nothing selected", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    await recordAnswerForMultipleChoice("token", "wb-1", "q-1", []);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        body: JSON.stringify({ selectedChoiceIds: [] }),
      }),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 400 });

    // when / then
    await expect(
      recordAnswerForMultipleChoice("token", "wb-1", "q-1", ["c1"]),
    ).rejects.toBeInstanceOf(Response);
  });
});
