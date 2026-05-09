import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { parseWordFillContent } from "~/components/workbook/schemas";
import type { StudyQuestionAudio } from "~/lib/api/study.server";

type WordFillCardProps = {
  content: string;
  audio?: StudyQuestionAudio;
  onAnswer: (correct: boolean) => void;
};

type Phase = "input" | "result";

function extractBlanks(text: string): { segments: string[]; answers: string[] } {
  const segments = text.split(/\{\{([^}]+)\}\}/g);
  const answers: string[] = [];
  for (let i = 1; i < segments.length; i += 2) {
    answers.push(segments[i]);
  }
  return { segments, answers };
}

function isCorrectAnswer(value: string, answer: string): boolean {
  return value.trim().toLowerCase() === answer.trim().toLowerCase();
}

// findNextUnlocked walks forward from `from` (exclusive) and returns the first
// blank whose `correct` flag is false, wrapping around. Returns -1 when every
// blank is already correct.
function findNextUnlocked(from: number, correct: boolean[]): number {
  for (let step = 1; step <= correct.length; step++) {
    const candidate = (from + step) % correct.length;
    if (!correct[candidate]) return candidate;
  }
  return -1;
}

// AudioPlayButton renders an <audio> control plus a small button that triggers
// playback. It is rendered only when a valid URL is provided so the play UI
// disappears entirely while the audio batch has not yet produced files.
function AudioPlayButton({ url, label }: { url: string; label: string }) {
  const { t } = useTranslation();
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const [failed, setFailed] = useState(false);

  function handlePlay() {
    const el = audioRef.current;
    if (!el) return;
    // Restart from the beginning so repeated taps always play from 0:00.
    el.currentTime = 0;
    setFailed(false);
    el.play().catch((reason) => {
      // Surface the failure as a small label so the user knows the tap landed
      // and that nothing will play. Common causes: autoplay rejection by the
      // browser, network error fetching the audio file, decode error.
      console.warn("audio playback failed", reason);
      setFailed(true);
    });
  }

  return (
    <div className="inline-flex flex-col items-end gap-1">
      <Button type="button" variant="outline" size="sm" aria-label={label} onClick={handlePlay}>
        <span aria-hidden="true">▶</span>
        <span className="ml-1 text-xs">{label}</span>
        {/* Inline audio element keeps the implementation self-contained; the
            element is hidden so the UI surface is just the button. */}
        <audio ref={audioRef} src={url} preload="none" className="hidden">
          <track kind="captions" />
        </audio>
      </Button>
      {failed && (
        <span role="alert" className="text-xs text-red-600">
          {t("workbooks.study.audioPlayFailed")}
        </span>
      )}
    </div>
  );
}

export function WordFillCard({ content, audio, onAnswer }: WordFillCardProps) {
  const { t } = useTranslation();
  const parsed = parseWordFillContent(content);
  const [inputs, setInputs] = useState<string[]>([]);
  const [phase, setPhase] = useState<Phase>("input");
  const inputRefs = useRef<HTMLInputElement[]>([]);
  const nextButtonRef = useRef<HTMLButtonElement | null>(null);

  // Focus the first blank when this question card mounts.
  useEffect(() => {
    inputRefs.current[0]?.focus();
  }, []);

  // Move focus to the Next button when transitioning to the result phase so the
  // user can advance with Enter/Space without reaching for the mouse.
  useEffect(() => {
    if (phase === "result") nextButtonRef.current?.focus();
  }, [phase]);

  if (!parsed?.target?.text) {
    return <p className="text-sm text-muted-foreground">{content}</p>;
  }

  const { segments, answers } = extractBlanks(parsed.target.text);

  if (inputs.length === 0 && answers.length > 0) {
    setInputs(new Array(answers.length).fill(""));
  }

  // A blank is "locked" once it holds the correct answer. Locked blanks become
  // read-only and are skipped by focus traversal — the user cannot accidentally
  // overwrite a value they have already gotten right.
  const correct = answers.map((answer, i) => isCorrectAnswer(inputs[i] ?? "", answer));
  const allCorrect = correct.length > 0 && correct.every(Boolean);
  const isResult = phase === "result";

  function handleInputChange(index: number, value: string) {
    if (correct[index]) return;

    const next = [...inputs];
    next[index] = value;
    setInputs(next);

    if (isResult) return;
    if (!isCorrectAnswer(value, answers[index])) return;

    const nextCorrect = answers.map((answer, i) => isCorrectAnswer(next[i] ?? "", answer));
    if (nextCorrect.every(Boolean)) {
      // Surface the result screen instead of advancing — the user reviews
      // their answer and explicitly continues via the Next button.
      setPhase("result");
      return;
    }

    // Move to the next blank that is still empty/wrong, wrapping around.
    const focusIndex = findNextUnlocked(index, nextCorrect);
    if (focusIndex >= 0) {
      inputRefs.current[focusIndex]?.focus();
      inputRefs.current[focusIndex]?.select();
    }
  }

  function handleReveal() {
    setPhase("result");
  }

  function handleNext() {
    onAnswer(allCorrect);
  }

  const sourceAudioUrl = audio?.source?.url ?? "";
  const targetAudioUrl = audio?.target?.url ?? "";

  return (
    <div className="space-y-6">
      {parsed.source?.text && (
        <div className="flex items-center gap-3 rounded-lg bg-muted/50 p-4">
          <p className="flex-1 text-lg font-medium">{parsed.source.text}</p>
          {sourceAudioUrl && (
            <AudioPlayButton url={sourceAudioUrl} label={t("workbooks.study.playSourceAudio")} />
          )}
        </div>
      )}

      <div className="flex flex-wrap items-center gap-1 text-lg">
        {segments.map((segment, i) => {
          if (i % 2 === 0) {
            return <span key={`text-${segment}`}>{segment}</span>;
          }
          const inputIndex = Math.floor(i / 2);
          const isLocked = correct[inputIndex];
          return (
            <span key={`blank-${inputIndex}`} className="inline-flex flex-col items-center">
              <Input
                ref={(el: HTMLInputElement | null) => {
                  if (el) inputRefs.current[inputIndex] = el;
                }}
                aria-label={t("workbooks.study.blankInput", { number: inputIndex + 1 })}
                className={`mx-1 inline-block w-32 text-center ${
                  isLocked
                    ? "border-green-500 bg-green-50 dark:bg-green-950/30"
                    : isResult
                      ? "border-red-500 bg-red-50 dark:bg-red-950/30"
                      : ""
                }`}
                value={isLocked ? answers[inputIndex] : (inputs[inputIndex] ?? "")}
                onChange={(e) => handleInputChange(inputIndex, e.target.value)}
                disabled={isResult || isLocked}
                readOnly={isLocked}
              />
              {isResult && !isLocked && (
                <span className="text-xs text-green-600">{answers[inputIndex]}</span>
              )}
            </span>
          );
        })}
      </div>

      {/* Target audio reveals the correct sentence. Held back until the user
          has finished the question to avoid leaking the answer. */}
      {isResult && targetAudioUrl && (
        <div className="flex justify-end">
          <AudioPlayButton url={targetAudioUrl} label={t("workbooks.study.playTargetAudio")} />
        </div>
      )}

      {parsed.explanation && isResult && (
        <p className="text-sm text-muted-foreground">{parsed.explanation}</p>
      )}

      {!isResult ? (
        <div className="flex justify-end">
          <Button onClick={handleReveal}>{t("workbooks.study.showAnswer")}</Button>
        </div>
      ) : (
        <div className="flex items-center justify-end gap-3">
          <span className={`text-sm font-medium ${allCorrect ? "text-green-600" : "text-red-600"}`}>
            {allCorrect ? t("workbooks.study.correct") : t("workbooks.study.incorrect")}
          </span>
          <Button ref={nextButtonRef} onClick={handleNext}>
            {t("workbooks.study.next")}
          </Button>
        </div>
      )}
    </div>
  );
}
