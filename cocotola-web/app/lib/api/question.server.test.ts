import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  addQuestion,
  deleteQuestion,
  listQuestions,
  type Question,
  QuestionValidationError,
  updateQuestion,
} from "./question.server";
import { catchError, jsonResponse, newClient, unauthorizedResponse } from "./test-helpers";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock);

const QUESTION_BASE_URL = "http://localhost:8090";

const BACKEND_MESSAGE = "source.lang must differ from target.lang";

const ADD_BODY = {
  questionType: "word_fill",
  content: '{"source":{"text":"ゴミを捨てる","lang":"ja"}}',
  tags: ["level:beginner"],
  orderIndex: 2,
};

const UPDATE_BODY = {
  content: '{"source":{"text":"ゴミを出す","lang":"ja"}}',
  tags: ["level:beginner"],
  orderIndex: 4,
};

function buildQuestion(overrides: Partial<Question> = {}): Question {
  return {
    questionId: "q-42",
    questionType: "word_fill",
    content: JSON.stringify({
      source: { text: "ゴミを捨てる", lang: "ja" },
      target: { text: "{{throw}} it {{away}}", lang: "en" },
    }),
    tags: ["level:beginner", "topic:phrasal-verbs"],
    orderIndex: 3,
    createdAt: "2026-02-03T04:05:06Z",
    updatedAt: "2026-05-06T07:08:09Z",
    ...overrides,
  };
}

function backendError(status: number): Response {
  return new Response(JSON.stringify({ message: BACKEND_MESSAGE }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
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

describe("listQuestions", () => {
  it("should return questions on success", async () => {
    // given
    const questions = [buildQuestion()];
    fetchMock.mockResolvedValue(jsonResponse({ questions }));

    // when
    const result = await listQuestions(newClient(), "wb-1");

    // then
    expect(result).toEqual(questions);
  });

  it("should send GET request with the access token", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ questions: [] }));

    // when
    await listQuestions(newClient(), "wb-1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-1/question`, {
      method: "GET",
      headers: { Authorization: "Bearer test-token" },
    });
  });

  it("should encode workbookId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ questions: [] }));

    // when
    await listQuestions(newClient(), "wb/1");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/question`,
      expect.any(Object),
    );
  });

  it("should return empty array when questions is undefined", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({}));

    // when
    const result = await listQuestions(newClient(), "wb-1");

    // then
    expect(result).toEqual([]);
  });

  it("should return empty array when questions is null", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ questions: null }));

    // when
    const result = await listQuestions(newClient(), "wb-1");

    // then
    expect(result).toEqual([]);
  });

  it("should accept a question without tags", async () => {
    // given
    const { tags: _, ...withoutTags } = buildQuestion();
    fetchMock.mockResolvedValue(jsonResponse({ questions: [withoutTags] }));

    // when
    const result = await listQuestions(newClient(), "wb-1");

    // then
    expect(result).toEqual([withoutTags]);
  });

  it("should accept a question that carries audio", async () => {
    // given
    const question = buildQuestion();
    const audio = {
      source: { url: "", durationSec: 0 },
      target: { url: "https://example.com/a.mp3", durationSec: 1.5 },
    };
    fetchMock.mockResolvedValue(jsonResponse({ questions: [{ ...question, audio }] }));

    // when
    const result = await listQuestions(newClient(), "wb-1");

    // then
    expect(result).toEqual([question]);
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 500 }));

    // when / then
    await expect(listQuestions(newClient(), "wb-1")).rejects.toMatchObject({ status: 500 });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(listQuestions(newClient(), "wb-1"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when a question in the response is malformed", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ questions: [{ questionId: "q-42" }] }));

    // when / then
    await expect(listQuestions(newClient(), "wb-1")).rejects.toMatchObject({ status: 502 });
  });
});

describe("addQuestion", () => {
  it("should return the added question on success", async () => {
    // given
    const question = buildQuestion({ questionId: "q-new" });
    fetchMock.mockResolvedValue(jsonResponse(question, 201));

    // when
    const result = await addQuestion(newClient(), "wb-1", ADD_BODY);

    // then
    expect(result).toEqual(question);
  });

  it("should send POST request with correct body", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildQuestion(), 201));

    // when
    await addQuestion(newClient(), "wb-1", ADD_BODY);

    // then
    expect(fetchMock).toHaveBeenCalledWith(`${QUESTION_BASE_URL}/api/v1/workbook/wb-1/question`, {
      method: "POST",
      headers: {
        Authorization: "Bearer test-token",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(ADD_BODY),
    });
  });

  it("should throw QuestionValidationError when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue(backendError(400));

    // when / then
    await expect(addQuestion(newClient(), "wb-1", ADD_BODY)).rejects.toBeInstanceOf(
      QuestionValidationError,
    );
  });

  it("should not include the backend body in QuestionValidationError", async () => {
    // given
    fetchMock.mockResolvedValue(backendError(400));

    // when
    const error = await catchError(addQuestion(newClient(), "wb-1", ADD_BODY));

    // then
    expect(String((error as Error).message)).not.toContain(BACKEND_MESSAGE);
  });

  it.each([403, 500])("should throw Response when API returns %i", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when / then
    await expect(addQuestion(newClient(), "wb-1", ADD_BODY)).rejects.toBeInstanceOf(Response);
  });

  it.each([403, 500])("should keep the status %i in the Response", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when / then
    await expect(addQuestion(newClient(), "wb-1", ADD_BODY)).rejects.toMatchObject({ status });
  });

  it.each([
    403, 500,
  ])("should not include the backend body in Response when API returns %i", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when
    const error = await catchError(addQuestion(newClient(), "wb-1", ADD_BODY));

    // then
    expect(await (error as Response).text()).not.toContain(BACKEND_MESSAGE);
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(addQuestion(newClient(), "wb-1", ADD_BODY));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response has a non-numeric orderIndex", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse({ ...buildQuestion(), orderIndex: "3" }, 201));

    // when / then
    await expect(addQuestion(newClient(), "wb-1", ADD_BODY)).rejects.toMatchObject({
      status: 502,
    });
  });
});

describe("updateQuestion", () => {
  it("should return the updated question on success", async () => {
    // given
    const question = buildQuestion({ orderIndex: 4 });
    fetchMock.mockResolvedValue(jsonResponse(question));

    // when
    const result = await updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY);

    // then
    expect(result).toEqual(question);
  });

  it("should send PUT request with correct URL and body", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildQuestion()));

    // when
    await updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/question/q-42`,
      {
        method: "PUT",
        headers: {
          Authorization: "Bearer test-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify(UPDATE_BODY),
      },
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(jsonResponse(buildQuestion()));

    // when
    await updateQuestion(newClient(), "wb/1", "q/2", UPDATE_BODY);

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/question/q%2F2`,
      expect.any(Object),
    );
  });

  it("should throw QuestionValidationError when API returns 400", async () => {
    // given
    fetchMock.mockResolvedValue(backendError(400));

    // when / then
    await expect(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY)).rejects.toBeInstanceOf(
      QuestionValidationError,
    );
  });

  it("should not include the backend body in QuestionValidationError", async () => {
    // given
    fetchMock.mockResolvedValue(backendError(400));

    // when
    const error = await catchError(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY));

    // then
    expect(String((error as Error).message)).not.toContain(BACKEND_MESSAGE);
  });

  it.each([403, 500])("should throw Response when API returns %i", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when / then
    await expect(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY)).rejects.toBeInstanceOf(
      Response,
    );
  });

  it.each([403, 500])("should keep the status %i in the Response", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when / then
    await expect(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY)).rejects.toMatchObject({
      status,
    });
  });

  it.each([
    403, 500,
  ])("should not include the backend body in Response when API returns %i", async (status) => {
    // given
    fetchMock.mockResolvedValue(backendError(status));

    // when
    const error = await catchError(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY));

    // then
    expect(await (error as Response).text()).not.toContain(BACKEND_MESSAGE);
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });

  it("should throw 502 Response when the response lacks content", async () => {
    // given
    const { content: _, ...withoutContent } = buildQuestion();
    fetchMock.mockResolvedValue(jsonResponse(withoutContent));

    // when / then
    await expect(updateQuestion(newClient(), "wb-1", "q-42", UPDATE_BODY)).rejects.toMatchObject({
      status: 502,
    });
  });
});

describe("deleteQuestion", () => {
  it("should send DELETE request with correct URL and headers", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteQuestion(newClient(), "wb-1", "q-42");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb-1/question/q-42`,
      {
        method: "DELETE",
        headers: { Authorization: "Bearer test-token" },
      },
    );
  });

  it("should encode workbookId and questionId in URL", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

    // when
    await deleteQuestion(newClient(), "wb/1", "q/2");

    // then
    expect(fetchMock).toHaveBeenCalledWith(
      `${QUESTION_BASE_URL}/api/v1/workbook/wb%2F1/question/q%2F2`,
      expect.any(Object),
    );
  });

  it("should throw Response when API returns error", async () => {
    // given
    fetchMock.mockResolvedValue(new Response(null, { status: 403 }));

    // when / then
    await expect(deleteQuestion(newClient(), "wb-1", "q-42")).rejects.toMatchObject({
      status: 403,
    });
  });

  it("should throw redirect to /login when API returns 401", async () => {
    // given
    fetchMock.mockResolvedValue(unauthorizedResponse());

    // when
    const caught = await catchError(deleteQuestion(newClient(), "wb-1", "q-42"));

    // then
    expect((caught as Response).headers.get("Location")).toBe("/login");
  });
});
