import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("~/lib/auth/require-auth.server", () => ({
  requireAuth: vi.fn(),
}));

vi.mock("~/lib/api/study.server", () => ({
  getStudyQuestions: vi.fn(),
  recordAnswer: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useLoaderData: vi.fn(),
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

import { useLoaderData } from "react-router";
import StudyPage from "./workbooks.$workbookId.study";

const mockedUseLoaderData = vi.mocked(useLoaderData);

describe("StudyPage", () => {
  it("should render empty state when no questions", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbookId: "wb-1", questions: [] });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("No questions due for study right now.")).toBeInTheDocument();
    const backLink = screen.getByText("Back to Workbook").closest("a");
    expect(backLink).toHaveAttribute("href", "/workbooks/wb-1");
  });

  it("should render page title and question count", () => {
    // given
    mockedUseLoaderData.mockReturnValue({ workbookId: "wb-1", questions: [] });

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
    mockedUseLoaderData.mockReturnValue({ workbookId: "wb-1", questions });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("1 / 1")).toBeInTheDocument();
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
    mockedUseLoaderData.mockReturnValue({ workbookId: "wb-1", questions });

    // when
    render(<StudyPage />);

    // then
    expect(screen.getByText("What is 1+1?")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });
});
