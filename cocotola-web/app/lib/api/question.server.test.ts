import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { addQuestion, deleteQuestion, listQuestions, updateQuestion } from "./question.server";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

describe("listQuestions", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should return questions on success", async () => {
    // given
    const questions = [{ questionId: "q-1", questionType: "word_fill", content: "{}" }];
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ questions }),
    });

    // when
    const result = await listQuestions("test-token", "wb-1");

    // then
    expect(result).toEqual(questions);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/wb-1/question", {
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should return empty array when questions is undefined", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    const result = await listQuestions("token", "wb-1");

    // then
    expect(result).toEqual([]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 500 });

    // when / then
    await expect(listQuestions("token", "wb-1")).rejects.toBeInstanceOf(Response);
  });
});

describe("addQuestion", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send POST request with correct body", async () => {
    // given
    const question = { questionId: "q-new" };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(question),
    });
    const body = { questionType: "word_fill", content: "{}", tags: ["lang:en"], orderIndex: 0 };

    // when
    const result = await addQuestion("test-token", "wb-1", body);

    // then
    expect(result).toEqual(question);
    expect(fetchMock).toHaveBeenCalledWith("http://localhost:8090/api/v1/workbook/wb-1/question", {
      method: "POST",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 400, text: () => Promise.resolve("bad") });

    // when / then
    await expect(
      addQuestion("token", "wb-1", { questionType: "word_fill", content: "{}", orderIndex: 0 }),
    ).rejects.toBeInstanceOf(Response);
  });
});

describe("updateQuestion", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send PUT request with correct URL and body", async () => {
    // given
    const question = { questionId: "q-1", content: "updated" };
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(question),
    });
    const body = { content: "updated", tags: ["lang:en"], orderIndex: 0 };

    // when
    const result = await updateQuestion("test-token", "wb-1", "q-1", body);

    // then
    expect(result).toEqual(question);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/question/q-1",
      {
        method: "PUT",
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
      },
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({}),
    });

    // when
    await updateQuestion("token", "wb/1", "q/2", { content: "{}", orderIndex: 0 });

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/question/q%2F2",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({
      ok: false,
      status: 403,
      text: () => Promise.resolve("forbidden"),
    });

    // when / then
    await expect(
      updateQuestion("token", "wb-1", "q-1", { content: "{}", orderIndex: 0 }),
    ).rejects.toBeInstanceOf(Response);
  });
});

describe("deleteQuestion", () => {
  beforeEach(() => {
    vi.stubEnv("QUESTION_BASE_URL", "http://localhost:8090");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("should send DELETE request with correct URL and headers", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await deleteQuestion("test-token", "wb-1", "q-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb-1/question/q-1",
      {
        method: "DELETE",
        headers: { Authorization: "Bearer test-token" },
      },
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: true });

    // when
    await deleteQuestion("token", "wb/1", "q/2");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8090/api/v1/workbook/wb%2F1/question/q%2F2",
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue({ ok: false, status: 403 });

    // when / then
    await expect(deleteQuestion("token", "wb-1", "q-1")).rejects.toBeInstanceOf(Response);
  });
});
