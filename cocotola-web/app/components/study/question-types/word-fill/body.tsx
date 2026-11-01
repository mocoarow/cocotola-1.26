import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import type { BodyProps } from "~/components/study/question-shell/types";
import { Input } from "~/components/ui/input";
import { extractBlanks, type WordFillContent } from "./answer";
import { AudioPlayButton } from "./audio-play-button";
import { useBlankInputs } from "./use-blank-inputs";

function blankClassName(isLocked: boolean, revealed: boolean): string {
  if (isLocked) return "border-green-500 bg-green-50 dark:bg-green-950/30";
  if (revealed) return "border-red-500 bg-red-50 dark:bg-red-950/30";
  return "";
}

type BlankFieldProps = {
  index: number;
  answer: string;
  value: string;
  isLocked: boolean;
  revealed: boolean;
  inputRef: (el: HTMLInputElement | null) => void;
  onChange: (value: string) => void;
};

function BlankField({
  index,
  answer,
  value,
  isLocked,
  revealed,
  inputRef,
  onChange,
}: BlankFieldProps) {
  const { t } = useTranslation();
  return (
    <span className="inline-flex flex-col items-center">
      <Input
        ref={inputRef}
        aria-label={t("workbooks.study.blankInput", { number: index + 1 })}
        className={`mx-1 inline-block w-32 text-center ${blankClassName(isLocked, revealed)}`}
        value={isLocked ? answer : value}
        onChange={(e) => onChange(e.target.value)}
        disabled={revealed || isLocked}
        readOnly={isLocked}
      />
      {revealed && !isLocked && <span className="text-xs text-green-600">{answer}</span>}
    </span>
  );
}

function SourcePrompt({ text, audioUrl }: { text: string; audioUrl: string }) {
  const { t } = useTranslation();
  return (
    <div className="flex items-center gap-3 rounded-lg bg-muted/50 p-4">
      <p className="flex-1 text-lg font-medium">{text}</p>
      {audioUrl && <AudioPlayButton url={audioUrl} label={t("workbooks.study.playSourceAudio")} />}
    </div>
  );
}

export function WordFillBody({
  content,
  audio,
  phase,
  onCommit,
}: BodyProps<"word_fill", WordFillContent>) {
  const { t } = useTranslation();
  const { segments, answers } = useMemo(
    () => extractBlanks(content.target.text),
    [content.target.text],
  );
  const revealed = phase === "revealed";
  const { inputs, correct, handleInputChange, registerInput } = useBlankInputs({
    answers,
    revealed,
    onCommit,
  });
  const targetAudioUrl = audio?.target?.url ?? "";

  return (
    <>
      {content.source?.text && (
        <SourcePrompt text={content.source.text} audioUrl={audio?.source?.url ?? ""} />
      )}

      <div className="flex flex-wrap items-center gap-1 text-lg">
        {segments.map((segment, i) => {
          if (i % 2 === 0) {
            return <span key={`text-${segment}`}>{segment}</span>;
          }
          const inputIndex = Math.floor(i / 2);
          return (
            <BlankField
              key={`blank-${inputIndex}`}
              index={inputIndex}
              answer={answers[inputIndex]}
              value={inputs[inputIndex] ?? ""}
              isLocked={correct[inputIndex]}
              revealed={revealed}
              inputRef={registerInput(inputIndex)}
              onChange={(value) => handleInputChange(inputIndex, value)}
            />
          );
        })}
      </div>

      {/* Held back until the answer is revealed to avoid leaking it. */}
      {revealed && targetAudioUrl && (
        <div className="flex justify-end">
          <AudioPlayButton url={targetAudioUrl} label={t("workbooks.study.playTargetAudio")} />
        </div>
      )}
    </>
  );
}
