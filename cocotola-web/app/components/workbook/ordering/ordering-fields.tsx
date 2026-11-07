import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "~/components/ui/input";
import { MAX_ORDERING_SEGMENTS, MAX_ORDERING_TEXT_BYTES } from "~/lib/study/answer-payload";
import { LangSelect } from "../lang-select";
import { MIN_ORDERING_CHIPS } from "../schemas";
import {
  type OrderingNotationError,
  type OrderingNotationResult,
  parseOrderingNotation,
} from "./notation";
import { OrderingPreview } from "./ordering-preview";

export type OrderingFieldsDefaults = {
  sourceText: string;
  sourceLang: string;
  targetLang: string;
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

function TargetFields({
  idPrefix,
  defaultLang,
  state,
}: {
  idPrefix: string;
  defaultLang: string;
  state: OrderingNotationState;
}) {
  const { t } = useTranslation();
  const [lang, setLang] = useState(defaultLang);
  return (
    <>
      <div className="space-y-1.5">
        <label htmlFor={`${idPrefix}targetLang`} className="text-sm font-medium">
          {t("workbooks.addQuestion.targetLang")}
        </label>
        <LangSelect
          id={`${idPrefix}targetLang`}
          name="targetLang"
          defaultValue={defaultLang}
          onChange={setLang}
        />
      </div>
      <NotationField id={`${idPrefix}targetText`} state={state} lang={lang} />
    </>
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
        aria-describedby={error ? errorId : undefined}
      />
      <p className="text-xs text-muted-foreground">
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
      <p className="text-xs text-muted-foreground">
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
  return (
    <div className="space-y-1.5">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <Input id={id} name={name} defaultValue={defaultValue} required={required} />
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

export function OrderingFields({
  idPrefix,
  defaults,
  state,
}: {
  idPrefix: string;
  defaults: OrderingFieldsDefaults;
  state: OrderingNotationState;
}) {
  const { t } = useTranslation();
  return (
    <>
      <div className="space-y-1.5">
        <label htmlFor={`${idPrefix}sourceLang`} className="text-sm font-medium">
          {t("workbooks.addQuestion.sourceLang")}
        </label>
        <LangSelect
          id={`${idPrefix}sourceLang`}
          name="sourceLang"
          defaultValue={defaults.sourceLang}
        />
      </div>
      <TextField
        id={`${idPrefix}sourceText`}
        name="sourceText"
        label={t("workbooks.addQuestion.sourceText")}
        defaultValue={defaults.sourceText}
        required
      />
      <TargetFields idPrefix={idPrefix} defaultLang={defaults.targetLang} state={state} />
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
