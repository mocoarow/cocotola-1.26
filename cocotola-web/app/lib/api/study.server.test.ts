import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getStudyQuestions, recordAnswer } from "./study.server";

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

describe("recordAnswer", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request with correct body for correct answer", async () => {
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
    const result = await recordAnswer("test-token", "wb-1", "q-1", true);

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
    await recordAnswer("token", "wb-1", "q-1", false);

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
    await recordAnswer("token", "wb/1", "q/2", true);

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
    await expect(recordAnswer("token", "wb-1", "q-1", true)).rejects.toBeInstanceOf(Response);
  });
});
