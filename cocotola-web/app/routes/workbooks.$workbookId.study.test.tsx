import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

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
    useNavigate: vi.fn(() => vi.fn()),
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

import { useLoaderData, useNavigate, useRouteLoaderData } from "react-router";
import { studySessionStorageKey } from "~/lib/study-session";
import StudyPage from "./workbooks.$workbookId.study";

const mockedUseLoaderData = vi.mocked(useLoaderData);
const mockedUseRouteLoaderData = vi.mocked(useRouteLoaderData);
const mockedUseNavigate = vi.mocked(useNavigate);

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
    practice: false,
    excludeIds: [],
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
  beforeEach(() => {
    // Reset resume-session storage so the in-progress set from a prior test
    // does not trigger a redirect in unrelated cases.
    window.localStorage.clear();
    // Reset loader-data mocks so a return value left over from one test (e.g.
    // the rerender path that calls mockReturnValue mid-test) cannot bleed
    // into a later test that forgets to call setLoaderData first.
    mockedUseLoaderData.mockReset();
    mockedUseRouteLoaderData.mockReset();
    // Keep a working default for useNavigate — most tests don't override it
    // and would otherwise see `undefined` and crash inside the resume effect.
    mockedUseNavigate.mockReturnValue(vi.fn());
  });

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

  // BaseUI Button with render={<Link/>} emits an <a href> but with role="button".
  // Lock the href + accessible name so future refactors don't silently drop the
  // practice CTA from the empty state. Match by getByText (text is stable across
  // the role override).
  it("empty state surfaces the Continue-practicing CTA pointing at ?practice=true", () => {
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
    const cta = screen.getByText("Continue practicing (no progress saved)").closest("a");
    expect(cta).toHaveAttribute("href", "/workbooks/wb-1/study?practice=true");
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

  it("should persist a session record on mount so a reload can resume", () => {
    // given: no prior session in localStorage
    window.localStorage.clear();
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

    // then: the session marker is saved with an empty answered set
    const raw = window.localStorage.getItem(
      studySessionStorageKey({ userId: "owner-1", workbookId: "wb-1", practice: false }),
    );
    expect(raw).not.toBeNull();
    const parsed = JSON.parse(raw as string);
    expect(parsed.userId).toBe("owner-1");
    expect(parsed.answeredIds).toEqual([]);
    expect(parsed.practice).toBe(false);
  });

  it("should not surface user A's answeredIds when user B opens the same workbook", () => {
    // Regression for the shared-browser scenario: A signs in, accumulates
    // answeredIds, signs out, B signs in. Without per-user scoping the
    // orchestrator would replay A's IDs into B's URL and queue.

    // given: user A's session is parked in localStorage under A's scope.
    window.localStorage.clear();
    window.localStorage.setItem(
      studySessionStorageKey({ userId: "user-a", workbookId: "wb-1", practice: false }),
      JSON.stringify({
        userId: "user-a",
        workbookId: "wb-1",
        practice: false,
        startedAtMs: Date.now(),
        answeredIds: ["q-secret-A1", "q-secret-A2"],
      }),
    );

    const navigateSpy = vi.fn();
    mockedUseNavigate.mockReturnValue(navigateSpy);

    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [
        {
          questionId: "q-1",
          questionType: "word_fill",
          content: JSON.stringify({
            source: { text: "hello", lang: "en" },
            target: { text: "{{hola}}", lang: "es" },
          }),
          orderIndex: 0,
        },
      ],
      currentUserId: "user-b",
    });

    // when: user B renders the study page on the same browser.
    render(<StudyPage />);

    // then: B sees nothing about A's resume state.
    expect(navigateSpy).not.toHaveBeenCalled();
    const bRaw = window.localStorage.getItem(
      studySessionStorageKey({ userId: "user-b", workbookId: "wb-1", practice: false }),
    );
    expect(bRaw).not.toBeNull();
    const bState = JSON.parse(bRaw as string);
    expect(bState.userId).toBe("user-b");
    expect(bState.answeredIds).toEqual([]);
    // And A's stash is left untouched — B never touched A's key.
    const aRaw = window.localStorage.getItem(
      studySessionStorageKey({ userId: "user-a", workbookId: "wb-1", practice: false }),
    );
    expect(aRaw).not.toBeNull();
    expect(JSON.parse(aRaw as string).answeredIds).toEqual(["q-secret-A1", "q-secret-A2"]);
  });

  it("should suppress the pre-resume answer card while a resume redirect is pending", () => {
    // Regression: between mount and the navigate-driven loader rerun, the
    // first card of the pre-exclusion queue used to be visible (and clickable
    // via fetcher.submit) for a moment. A user racing the redirect could
    // submit an answer for a question they had already finished, sending the
    // SRS engine a second update.

    // given: localStorage says q-1 is already answered, but the URL carries
    // no excludeIds and the loader returned the full pool starting with q-1.
    window.localStorage.clear();
    window.localStorage.setItem(
      studySessionStorageKey({ userId: "owner-1", workbookId: "wb-1", practice: false }),
      JSON.stringify({
        userId: "owner-1",
        workbookId: "wb-1",
        practice: false,
        startedAtMs: Date.now(),
        answeredIds: ["q-1"],
      }),
    );
    mockedUseNavigate.mockReturnValue(vi.fn());

    const wordFill = (text: string) =>
      JSON.stringify({
        source: { text, lang: "en" },
        target: { text: "{{translated}}", lang: "es" },
      });

    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [
        { questionId: "q-1", questionType: "word_fill", content: wordFill("ALPHA"), orderIndex: 0 },
        { questionId: "q-2", questionType: "word_fill", content: wordFill("BETA"), orderIndex: 1 },
      ],
      currentUserId: "owner-1",
    });

    // when: mount fires the resume layout effect synchronously, flips the
    // gate to "redirecting", and renders the placeholder instead of the
    // answer card.
    render(<StudyPage />);

    // then: the placeholder is mounted; the pre-resume card is not.
    expect(screen.getByTestId("study-resume-pending")).toBeInTheDocument();
    expect(screen.queryByText("ALPHA")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Check" })).not.toBeInTheDocument();
  });

  it("should remount and show the first unanswered question after a resume-driven navigate", () => {
    // Regression: the lazy useState seed in StudySession only runs once per
    // mount. Without a key change on excludeIds, navigate would trigger a
    // loader rerun, the loader's post-exclusion list would arrive via
    // useLoaderData, but the queue would still hold the pre-exclusion items
    // and re-show questions the user already answered.

    // given: localStorage says q-1, q-2 are already answered, but the URL
    // still carries no excludeIds and the first loader call returned the full
    // 3-question pool.
    window.localStorage.clear();
    window.localStorage.setItem(
      studySessionStorageKey({ userId: "owner-1", workbookId: "wb-1", practice: false }),
      JSON.stringify({
        userId: "owner-1",
        workbookId: "wb-1",
        practice: false,
        startedAtMs: Date.now(),
        answeredIds: ["q-1", "q-2"],
      }),
    );
    mockedUseNavigate.mockReturnValue(vi.fn());

    const wordFill = (text: string) =>
      JSON.stringify({
        source: { text, lang: "en" },
        target: { text: "{{translated}}", lang: "es" },
      });

    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [
        { questionId: "q-1", questionType: "word_fill", content: wordFill("ALPHA"), orderIndex: 0 },
        { questionId: "q-2", questionType: "word_fill", content: wordFill("BETA"), orderIndex: 1 },
        { questionId: "q-3", questionType: "word_fill", content: wordFill("GAMMA"), orderIndex: 2 },
      ],
      currentUserId: "owner-1",
    });

    // when: render with the first loader payload (mount fires the resume
    // navigate inside an effect)
    const { rerender } = render(<StudyPage />);

    // and: simulate the loader rerun delivering the post-exclusion list with
    // the new excludeIds reflected in the URL.
    mockedUseLoaderData.mockReturnValue({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [
        { questionId: "q-3", questionType: "word_fill", content: wordFill("GAMMA"), orderIndex: 2 },
      ],
      practice: false,
      excludeIds: ["q-1", "q-2"],
    });
    rerender(<StudyPage />);

    // then: StudySession was remounted by the keyed transition and the
    // visible card is GAMMA (q-3), not the stale ALPHA (q-1).
    expect(screen.getByText("GAMMA")).toBeInTheDocument();
    expect(screen.queryByText("ALPHA")).not.toBeInTheDocument();
  });

  it("should redirect with excludeIds query params when a session has answered ids", () => {
    // given: a fresh session with one already-answered question
    window.localStorage.clear();
    window.localStorage.setItem(
      studySessionStorageKey({ userId: "owner-1", workbookId: "wb-1", practice: false }),
      JSON.stringify({
        userId: "owner-1",
        workbookId: "wb-1",
        practice: false,
        startedAtMs: Date.now(),
        answeredIds: ["q-1"],
      }),
    );

    const navigateSpy = vi.fn();
    mockedUseNavigate.mockReturnValue(navigateSpy);

    setLoaderData({
      workbookId: "wb-1",
      workbookOwnerId: "owner-1",
      questions: [
        {
          questionId: "q-2",
          questionType: "word_fill",
          content: JSON.stringify({
            source: { text: "hi", lang: "en" },
            target: { text: "{{hola}}", lang: "es" },
          }),
          orderIndex: 0,
        },
      ],
      currentUserId: "owner-1",
    });

    // when
    render(<StudyPage />);

    // then: navigate was called with excludeIds in the search string
    expect(navigateSpy).toHaveBeenCalledTimes(1);
    const target = navigateSpy.mock.calls[0]?.[0];
    expect(typeof target).toBe("string");
    expect(target as string).toMatch(/excludeIds=q-1/);
  });

  it("should clear the localStorage session when the user finishes the last question", async () => {
    // Verifies that finishing the queue removes the resume marker — without
    // putting the clear call inside the setQueue updater (which would violate
    // the React rule that updaters be pure).

    // given
    window.localStorage.clear();
    const user = userEvent.setup();
    const questions = [
      {
        questionId: "q-1",
        questionType: "multiple_choice",
        content: JSON.stringify({
          questionText: "Only question",
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
    // The mount effect seeds an empty session.
    const sessionKey = studySessionStorageKey({
      userId: "owner-1",
      workbookId: "wb-1",
      practice: false,
    });
    expect(window.localStorage.getItem(sessionKey)).not.toBeNull();

    // when: answer correctly, advance to the done screen.
    await user.click(screen.getByText("Right"));
    await user.click(screen.getByRole("button", { name: "Check" }));
    await user.click(screen.getByRole("button", { name: "Next" }));

    // then: session marker is gone.
    expect(screen.getByText("Session Complete!")).toBeInTheDocument();
    expect(window.localStorage.getItem(sessionKey)).toBeNull();
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
