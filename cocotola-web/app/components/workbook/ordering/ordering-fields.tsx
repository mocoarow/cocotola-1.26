import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "~/components/ui/input";
import { MAX_ORDERING_SEGMENTS, MAX_ORDERING_TEXT_BYTES } from "~/lib/study/answer-payload";
import { type LangPair, SourceLangField, TargetLangField } from "../lang-pair";
import { MIN_ORDERING_CHIPS } from "../schemas";
import {
  type OrderingNotationError,
  type OrderingNotationResult,
  parseOrderingNotation,
} from "./notation";
import { OrderingPreview } from "./ordering-preview";

export type OrderingFieldsDefaults = {
  sourceText: string;
  explanation: string;
  tags: string;
};

export type OrderingNotationState = {
  notation: string;
  setNotation: (notation: string) => void;
  result: OrderingNotationResult;
};

const ERROR_PARAMS: Partial<Record<OrderingNotationError, Record<string, number>>> = {
  tooFewChips: { min: MIN_ORDERING_CHIPS },
  tooManySegments: { max: MAX_ORDERING_SEGMENTS },
  textTooLong: { max: MAX_ORDERING_TEXT_BYTES },
};

export function useOrderingNotation(initial: string): OrderingNotationState {
  const [notation, setNotation] = useState(initial);
  return { notation, setNotation, result: parseOrderingNotation(notation) };
}

function shownError(state: OrderingNotationState): OrderingNotationError | null {
  if (state.result.ok || !state.notation.trim()) return null;
  return state.result.error;
}

function NotationError({ id, error }: { id: string; error: OrderingNotationError | null }) {
  const { t } = useTranslation();
  return (
    <p id={id} aria-live="polite" className="text-xs text-destructive">
      {error && t(`workbooks.addQuestion.orderingErrors.${error}`, ERROR_PARAMS[error])}
    </p>
  );
}

function NotationField({
  id,
  state,
  lang,
}: {
  id: string;
  state: OrderingNotationState;
  lang: string;
}) {
  const { t } = useTranslation();
  const error = shownError(state);
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;
  const noteId = `${id}-note`;
  const describedBy = error ? `${hintId} ${errorId} ${noteId}` : `${hintId} ${noteId}`;
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {t("workbooks.addQuestion.orderingTarget")}
      </label>
      <Input
        id={id}
        name="targetText"
        value={state.notation}
        onChange={(e) => state.setNotation(e.target.value)}
        placeholder="I / am / a student [[.]]"
        required
        aria-invalid={error !== null}
        aria-describedby={describedBy}
      />
      <p id={hintId} className="text-xs text-muted-foreground">
        {t("workbooks.addQuestion.orderingTargetHint")}
      </p>
      <NotationError id={errorId} error={error} />
      {state.result.ok && (
        <OrderingPreview
          segments={state.result.segments}
          lang={lang}
          label={t("workbooks.addQuestion.orderingPreview")}
        />
      )}
      <p id={noteId} className="text-xs text-muted-foreground">
        {t("workbooks.addQuestion.orderingCapitalizationNote")}
      </p>
    </div>
  );
}

function TextField({
  id,
  name,
  label,
  defaultValue,
  required,
  hint,
}: {
  id: string;
  name: string;
  label: string;
  defaultValue: string;
  required?: boolean;
  hint?: string;
}) {
  const hintId = `${id}-hint`;
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <Input
        id={id}
        name={name}
        defaultValue={defaultValue}
        required={required}
        aria-describedby={hint ? hintId : undefined}
      />
      {hint && (
        <p id={hintId} className="text-xs text-muted-foreground">
          {hint}
        </p>
      )}
    </div>
  );
}

export function OrderingFields({
  idPrefix,
  defaults,
  langs,
  state,
}: {
  idPrefix: string;
  defaults: OrderingFieldsDefaults;
  langs: LangPair;
  state: OrderingNotationState;
}) {
  const { t } = useTranslation();
  return (
    <>
      <SourceLangField id={`${idPrefix}sourceLang`} langs={langs} />
      <TextField
        id={`${idPrefix}sourceText`}
        name="sourceText"
        label={t("workbooks.addQuestion.sourceText")}
        defaultValue={defaults.sourceText}
        required
      />
      <TargetLangField id={`${idPrefix}targetLang`} langs={langs} />
      <NotationField id={`${idPrefix}targetText`} state={state} lang={langs.targetLang} />
      <TextField
        id={`${idPrefix}explanation`}
        name="explanation"
        label={t("workbooks.addQuestion.explanation")}
        defaultValue={defaults.explanation}
      />
      <TextField
        id={`${idPrefix}tags`}
        name="tags"
        label={t("workbooks.addQuestion.tags")}
        defaultValue={defaults.tags}
        hint={t("workbooks.addQuestion.tagsHint")}
      />
    </>
  );
}
