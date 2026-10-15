import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { MarkdownText } from "./markdown-text";

const SENTENCE_URL = "https://tatoeba.org/en/sentences/show/123";
const SENTENCE_LINK = `Tatoeba [#123](${SENTENCE_URL})`;

describe("MarkdownText", () => {
  it("should render a markdown link as an anchor with the given href", () => {
    // given
    const text = SENTENCE_LINK;

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByRole("link", { name: "#123" })).toHaveAttribute("href", SENTENCE_URL);
  });

  it("should open links in a new tab", () => {
    // given
    const text = SENTENCE_LINK;

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByRole("link", { name: "#123" })).toHaveAttribute("target", "_blank");
  });

  it("should set rel noopener noreferrer on links", () => {
    // given
    const text = SENTENCE_LINK;

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByRole("link", { name: "#123" })).toHaveAttribute(
      "rel",
      "noopener noreferrer",
    );
  });

  it("should not keep a javascript: href when the link uses the javascript scheme", () => {
    // given
    const text = "[click](javascript:alert(1))";

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByText("click").closest("a")).not.toHaveAttribute(
      "href",
      expect.stringMatching(/^javascript:/i),
    );
  });

  it("should not render a script element when raw html is given", () => {
    // given
    const text = "<script>alert(1)</script>";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("script")).toBeNull();
  });

  it("should not render an img element when the text contains an image", () => {
    // given
    const text = "![pixel](https://example.test/p.gif)";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("img")).toBeNull();
  });

  it("should render a heading when the text contains a heading", () => {
    // given
    const text = "# Title";

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByRole("heading", { name: "Title" })).toBeInTheDocument();
  });

  it("should render an ordered list when the text starts with a number marker", () => {
    // given
    const text = "1. first";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("ol > li")).toHaveTextContent("first");
  });

  it("should render an unordered list when the text starts with a hyphen marker", () => {
    // given
    const text = "- note";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("ul > li")).toHaveTextContent("note");
  });

  it("should render a blockquote when the text starts with a quote marker", () => {
    // given
    const text = "> aside";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("blockquote")).toHaveTextContent("aside");
  });

  it("should render a horizontal rule when the text contains a thematic break", () => {
    // given
    const text = "above\n\n---\n\nbelow";

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByRole("separator")).toBeInTheDocument();
  });

  it("should render a code block when the text contains a fenced code block", () => {
    // given
    const text = "```\nconst x = 1;\n```";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("pre > code")).toHaveTextContent("const x = 1;");
  });

  it("should render a line break as br when the text contains a single newline", () => {
    // given
    const text = "first line\nsecond line";

    // when
    const { container } = render(<MarkdownText text={text} />);

    // then
    expect(container.querySelector("br")).not.toBeNull();
  });

  it("should apply className to the wrapper element", () => {
    // given
    const className = "text-sm";

    // when
    const { container } = render(<MarkdownText text="plain" className={className} />);

    // then
    expect(container.firstChild).toHaveClass(className);
  });

  it("should render plain text unchanged when no markdown syntax is present", () => {
    // given
    const text = "Because reasons";

    // when
    render(<MarkdownText text={text} />);

    // then
    expect(screen.getByText(text)).toBeInTheDocument();
  });
});
