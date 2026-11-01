import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  deleteStudyHistory,
  getStudyQuestions,
  getStudySummary,
  listStudyRecords,
  recordAnswer,
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

  it("should append repeated excludeIds query params", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions: [], totalDue: 0, newCount: 0, reviewCount: 0 }),
    });

    // when
    await getStudyQuestions("token", "wb-1", 10, false, ["q-1", "q-2"]);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study?limit=10&excludeIds=q-1&excludeIds=q-2",
      expect.any(Object),
    );
  });

  it("should skip empty-string entries in excludeIds", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions: [], totalDue: 0, newCount: 0, reviewCount: 0 }),
    });

    // when
    await getStudyQuestions("token", "wb-1", 10, false, ["", "q-1"]);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study?limit=10&excludeIds=q-1",
      expect.any(Object),
    );
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

describe("recordAnswer", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request with correct=true body when payload is word_fill", async () => {
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
    const result = await recordAnswer("test-token", "wb-1", "q-1", {
      type: "word_fill",
      correct: true,
    });

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

  it("should send correct=false when word_fill answer is incorrect", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    await recordAnswer("token", "wb-1", "q-1", { type: "word_fill", correct: false });

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
    await recordAnswer("token", "wb/1", "q/2", { type: "word_fill", correct: true });

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
    await expect(
      recordAnswer("token", "wb-1", "q-1", { type: "word_fill", correct: true }),
    ).rejects.toBeInstanceOf(Response);
  });

  it("should send POST request with selectedChoiceIds body when payload is multiple_choice", async () => {
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
    const result = await recordAnswer("test-token", "wb-1", "q-1", {
      type: "multiple_choice",
      selectedChoiceIds: ["c1", "c2"],
    });

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
    await recordAnswer("token", "wb-1", "q-1", { type: "multiple_choice", selectedChoiceIds: [] });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        body: JSON.stringify({ selectedChoiceIds: [] }),
      }),
    );
  });
});

describe("deleteStudyHistory", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send DELETE request and resolve on 204", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true, status: 204 });

    // when
    await deleteStudyHistory("test-token", "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/wb-1/study", {
      method: "DELETE",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true, status: 204 });

    // when
    await deleteStudyHistory("token", "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/study",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(deleteStudyHistory("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});

describe("listStudyRecords", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return records on success", async () => {
    // given
    const response = {
      records: [
        {
          workbookId: "wb-1",
          questionId: "q-1",
          consecutiveCorrect: 2,
          lastAnsweredAt: "2026-04-25T10:00:00Z",
          nextDueAt: "2026-04-26T10:00:00Z",
          totalCorrect: 3,
          totalIncorrect: 1,
        },
      ],
    };
    fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve(response) });

    // when
    const result = await listStudyRecords("test-token", "wb-1");

    // then
    expect(result).toEqual(response);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/study/records",
      { headers: { Authorization: "Bearer test-token" } },
    );
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true, json: () => Promise.resolve({ records: [] }) });

    // when
    await listStudyRecords("token", "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/study/records",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(listStudyRecords("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});
