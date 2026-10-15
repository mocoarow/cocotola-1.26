import type { ComponentPropsWithoutRef } from "react";
import Markdown, { type ExtraProps } from "react-markdown";
import remarkBreaks from "remark-breaks";

import { cn } from "~/lib/utils";

// img is excluded so that rendering an explanation never fetches a remote resource.
const ALLOWED_ELEMENTS = [
  "p",
  "a",
  "strong",
  "em",
  "code",
  "pre",
  "br",
  "h1",
  "h2",
  "h3",
  "h4",
  "h5",
  "h6",
  "ul",
  "ol",
  "li",
  "blockquote",
  "hr",
];
const REMARK_PLUGINS = [remarkBreaks];

// Tailwind preflight strips heading sizes and list markers, so restore them here.
const BLOCK_STYLES = [
  "space-y-2",
  "[&_:is(h1,h2,h3,h4,h5,h6)]:font-semibold",
  "[&_h1]:text-lg",
  "[&_h2]:text-base",
  "[&_ul]:list-disc",
  "[&_ol]:list-decimal",
  "[&_:is(ul,ol)]:pl-5",
  "[&_blockquote]:border-l-2",
  "[&_blockquote]:pl-3",
  "[&_pre]:overflow-x-auto",
  "[&_pre]:rounded",
  "[&_pre]:bg-muted",
  "[&_pre]:p-2",
];

function ExternalLink({ node: _node, ...props }: ComponentPropsWithoutRef<"a"> & ExtraProps) {
  return (
    <a
      {...props}
      target="_blank"
      rel="noopener noreferrer"
      className="underline underline-offset-2 hover:text-foreground"
    />
  );
}

const COMPONENTS = { a: ExternalLink };

type MarkdownTextProps = {
  text: string;
  className?: string;
};

export function MarkdownText({ text, className }: MarkdownTextProps) {
  return (
    <div className={cn(BLOCK_STYLES, className)}>
      <Markdown
        remarkPlugins={REMARK_PLUGINS}
        allowedElements={ALLOWED_ELEMENTS}
        components={COMPONENTS}
      >
        {text}
      </Markdown>
    </div>
  );
}
