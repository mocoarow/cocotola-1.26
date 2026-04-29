import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/study.server", () => ({
  getStudyQuestions: vi.fn(),
  recordAnswerForMultipleChoice: vi.fn(),
  recordAnswerForWordFill: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
    useRouteLoaderData: vi.fn(),
    useFetcher: vi.fn(() => ({
      state: "idle",
      submit: vi.fn(),
      Form: "form",
    })),
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { useLoaderData, useRouteLoaderData } from "react-router";
import StudyPage from "./workbooks.$workbookId.study";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedUseRouteLoaderData = vi.mocked(useRouteLoaderData);

function setLoaderData(args: {
  workbookId: string;
  workbookOwnerId: string;
  questions: Array<{
    questionId: string;
    questionType: string;
    content: string;
    orderIndex: number;
  }>;
  currentUserId?: string | null;
}) {
  mockedUseLoaderData.mockReturnValue({
    workbookId: args.workbookId,
    workbookOwnerId: args.workbookOwnerId,
    questions: args.questions,
  });
  mockedUseRouteLoaderData.mockReturnValue(
    args.currentUserId === null
      ? { user: null }
      : {
          user: {
            userId: args.currentUserId ?? args.workbookOwnerId,
            loginId: "u",
            organizationName: "o",
          },
        },
  );
}

describe("StudyPage", () => {
  it("should render empty state when no questions", () => {
    // given
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [],
      currentUserId: "owner-1",
    });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("No questions due for study right now.")).toBeInTheDocument();
    const backLink = screen.getByText("Back to Workbook").closest("a");
    expect(backLink).toHaveAttribute("href", "/workbooks/wb-1");
  });

  it("should send empty-state back link to public list when current user is not the owner", () => {
    // given
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-other",
      questions: [],
      currentUserId: "viewer-1",
    });

    // when
    render(<StudyPage />);

    // then
    expect(screen.queryByText("Back to Workbook")).not.toBeInTheDocument();
    const backLink = screen.getByText("Back to Public Workbooks").closest("a");
    expect(backLink).toHaveAttribute("href", "/workbooks/public");
  });

  it("should treat unauthenticated session (no user) as non-owner", () => {
    // given
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [],
      currentUserId: null,
    });

    // when
    render(<StudyPage />);

    // then
    const backLink = screen.getByText("Back to Public Workbooks").closest("a");
    expect(backLink).toHaveAttribute("href", "/workbooks/public");
  });

  it("should render page title and question count", () => {
    // given
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [],
      currentUserId: "owner-1",
    });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("Study Session")).toBeInTheDocument();
    expect(screen.getByText("0 questions to study")).toBeInTheDocument();
  });

  it("should render word fill card for word_fill question type", () => {
    // given
    const questions = [
      {
        questionId: "q-1",
        questionType: "word_fill",
        content: JSON.stringify({
          source: { text: "hello", lang: "en" },
          target: { text: "{{hola}}", lang: "es" },
        }),
        orderIndex: 0,
      },
    ];
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions,
      currentUserId: "owner-1",
    });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("0 / 1")).toBeInTheDocument();
  });

  it("should render multiple choice card for multiple_choice question type", () => {
    // given
    const questions = [
      {
        questionId: "q-1",
        questionType: "multiple_choice",
        content: JSON.stringify({
          questionText: "What is 1+1?",
          choices: [
            { id: "1", text: "2", isCorrect: true },
            { id: "2", text: "3", isCorrect: false },
          ],
        }),
        orderIndex: 0,
      },
    ];
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions,
      currentUserId: "owner-1",
    });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("What is 1+1?")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("should re-present a wrong-answered question and not finish the session until it is correct", async () => {
    // given
    const user = userEvent.setup();
    const questions = [
      {
        questionId: "q-1",
        questionType: "multiple_choice",
        content: JSON.stringify({
          questionText: "Pick the right one",
          choices: [
            { id: "right", text: "Right", isCorrect: true },
            { id: "wrong", text: "Wrong", isCorrect: false },
          ],
          shuffleChoices: false,
          showCorrectCount: false,
        }),
        orderIndex: 0,
      },
    ];
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions,
      currentUserId: "owner-1",
    });
    render(<StudyPage />);

    // when (answer wrong)
    await user.click(screen.getByText("Wrong"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then (same question reappears with reset state, session not done)
    expect(screen.getByText("Pick the right one")).toBeInTheDocument();
    expect(screen.queryByText("Incorrect")).not.toBeInTheDocument();
    expect(screen.queryByText("Session Complete!")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Check" })).toBeDisabled();

    // when (answer correctly)
    await user.click(screen.getByText("Right"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then (session complete)
    expect(screen.getByText("Session Complete!")).toBeInTheDocument();
  });

  it("should send wrong-answered question to the back of the queue and require all to be correct", async () => {
    // given
    const user = userEvent.setup();
    const questions = [
      {
        questionId: "q-1",
        questionType: "multiple_choice",
        content: JSON.stringify({
          questionText: "Q1?",
          choices: [
            { id: "1a", text: "Q1-Right", isCorrect: true },
            { id: "1b", text: "Q1-Wrong", isCorrect: false },
          ],
          shuffleChoices: false,
          showCorrectCount: false,
        }),
        orderIndex: 0,
      },
      {
        questionId: "q-2",
        questionType: "multiple_choice",
        content: JSON.stringify({
          questionText: "Q2?",
          choices: [
            { id: "2a", text: "Q2-Right", isCorrect: true },
            { id: "2b", text: "Q2-Wrong", isCorrect: false },
          ],
          shuffleChoices: false,
          showCorrectCount: false,
        }),
        orderIndex: 1,
      },
    ];
    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions,
      currentUserId: "owner-1",
    });
    render(<StudyPage />);

    // when (Q1 wrong → goes to back; Q2 should be next)
    await user.click(screen.getByText("Q1-Wrong"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then (Q2 is now visible)
    expect(screen.getByText("Q2?")).toBeInTheDocument();
    expect(screen.queryByText("Q1?")).not.toBeInTheDocument();

    // when (Q2 correct → Q1 should re-appear)
    await user.click(screen.getByText("Q2-Right"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then (Q1 reappears, session not complete yet)
    expect(screen.getByText("Q1?")).toBeInTheDocument();
    expect(screen.queryByText("Session Complete!")).not.toBeInTheDocument();

    // when (Q1 correct)
    await user.click(screen.getByText("Q1-Right"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then (session complete)
    expect(screen.getByText("Session Complete!")).toBeInTheDocument();
  });
});
