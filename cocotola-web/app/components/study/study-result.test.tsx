import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import "~/i18n/config";

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});

import { StudyResult } from "./study-result";

describe("StudyResult", () => {
  it("should render score percentage", () => {
    // given / when
    render(
      <StudyResult
        correctCount={7}
        incorrectCount={3}
        backUrl="/workbooks/wb-1"
        backLabel="Back to Workbook"
      />,
    );

    // then
    expect(screen.getByText("You scored 70%")).toBeInTheDocument();
  });

  it("should render correct and incorrect counts", () => {
    // given / when
    render(
      <StudyResult
        correctCount={7}
        incorrectCount={3}
        backUrl="/workbooks/wb-1"
        backLabel="Back to Workbook"
      />,
    );

    // then
    expect(screen.getByText("7")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.getByText("10")).toBeInTheDocument();
  });

  it("should render the provided back link with the provided label and url", () => {
    // given / when
    render(
      <StudyResult
        correctCount={0}
        incorrectCount={0}
        backUrl="/workbooks/wb-42"
        backLabel="Back to Workbook"
      />,
    );

    // then
    const link = screen.getByText("Back to Workbook").closest("a");
    expect(link).toHaveAttribute("href", "/workbooks/wb-42");
  });

  it("should render the public list back link when given the public destination", () => {
    // given / when
    render(
      <StudyResult
        correctCount={0}
        incorrectCount={0}
        backUrl="/workbooks/public"
        backLabel="Back to Public Workbooks"
      />,
    );

    // then
    const link = screen.getByText("Back to Public Workbooks").closest("a");
    expect(link).toHaveAttribute("href", "/workbooks/public");
  });

  it("should render 0% when no questions answered", () => {
    // given / when
    render(
      <StudyResult
        correctCount={0}
        incorrectCount={0}
        backUrl="/workbooks/wb-1"
        backLabel="Back to Workbook"
      />,
    );

    // then
    expect(screen.getByText("You scored 0%")).toBeInTheDocument();
  });

  it("should render session complete title", () => {
    // given / when
    render(
      <StudyResult
        correctCount={5}
        incorrectCount={5}
        backUrl="/workbooks/wb-1"
        backLabel="Back to Workbook"
      />,
    );

    // then
    expect(screen.getByText("Session Complete!")).toBeInTheDocument();
  });
});
