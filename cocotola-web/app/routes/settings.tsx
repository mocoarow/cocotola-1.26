import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Form, useActionData, useFetcher, useLoaderData, useNavigation } from "react-router";

import { Button } from "~/components/ui/button";
import { type SupportedLanguage, supportedLanguages } from "~/i18n/config";
import {
  getUserPreferences,
  updateUserDailyGoal,
  updateUserLanguage,
  updateUserTimezone,
} from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { detectBrowserTimezone, isValidIanaTimezoneShape } from "~/lib/format/local-date";
import { MAX_DAILY_GOAL, MIN_DAILY_GOAL } from "~/lib/format/user-setting-bounds";

import type { Route } from "./+types/settings";

/**
 * Server loader: hydrates the form with the persisted preferences from
 * /auth/me. Funneled through getUserPreferences so the response shape and
 * the 401-redirect-to-login contract stay in sync with routes/workbooks.tsx
 * and routes/dashboard.tsx.
 */
export async function loader({ request }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);
  const prefs = await getUserPreferences(request, accessToken);
  return {
    loginId: prefs.loginId,
    organizationName: prefs.organizationName,
    language: prefs.language,
    dailyGoal: prefs.dailyGoal,
    timezone: prefs.timezone,
  };
}

// Generic body returned for any tampered submit. The specific cause
// (unknown intent, dailyGoal out of range, timezone shape, unsupported
// language) is recorded server-side instead — the legitimate form paths
// never trigger these branches.
const REJECTED_REQUEST_BODY = "Invalid request parameters";

/**
 * Saves a single preference field. Catches per-field failures so the
 * caller can report partial success while still propagating thrown
 * redirect Responses (e.g. the /login redirect from redirectOnUnauthorized)
 * so React Router can navigate instead of recording the redirect as a
 * field-level save failure.
 */
async function savePreferenceField<F extends string>(
  field: F,
  saveFn: () => Promise<void>,
): Promise<{ field: F; ok: boolean }> {
  try {
    await saveFn();
    return { field, ok: true };
  } catch (err) {
    if (err instanceof Response && err.status >= 300 && err.status < 400) {
      throw err;
    }
    console.error(`[settings] save ${field} failed`, err);
    return { field, ok: false };
  }
}

/**
 * Handles intent=saveLanguage. Validates the submitted code against the
 * allow-list before forwarding to the backend — the client's UI cannot
 * produce an unsupported value, so reaching this branch is always a
 * tampered submit and is logged with a generic length-only fingerprint.
 *
 * Returns the persisted language in the response so the client can
 * commit i18n.changeLanguage only after the server confirms the write
 * (the failure path leaves the UI on the pre-change language; see
 * SettingsPage's languageFetcher effect).
 */
async function handleLanguageIntent(request: Request, accessToken: string, formData: FormData) {
  const language = String(formData.get("language") ?? "");
  if (!supportedLanguages.includes(language as SupportedLanguage)) {
    console.warn(`[settings] action rejected: unsupported language (length=${language.length})`);
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }
  await updateUserLanguage(request, accessToken, language);
  return { ok: true as const, intent: "saveLanguage" as const, language };
}

/**
 * Handles intent=saveStudyPreferences. Validates dailyGoal + timezone,
 * then saves each field sequentially so a per-field failure can be
 * reported back to the UI without hiding the other field's success.
 */
async function handleStudyIntent(request: Request, accessToken: string, formData: FormData) {
  const dailyGoalRaw = formData.get("dailyGoal");
  const timezone = String(formData.get("timezone") ?? "");

  // Require a digit-only string. Number.parseInt accepts partial parses
  // ("30abc" => 30), letting tampered form bodies sneak past the range
  // check; the /^\d+$/ shape guard converts every malformed input to NaN.
  const dailyGoal =
    typeof dailyGoalRaw === "string" && /^\d+$/.test(dailyGoalRaw)
      ? Number.parseInt(dailyGoalRaw, 10)
      : Number.NaN;
  if (!Number.isFinite(dailyGoal) || dailyGoal < MIN_DAILY_GOAL || dailyGoal > MAX_DAILY_GOAL) {
    console.warn(`[settings] action rejected: dailyGoal out of range (got ${dailyGoal})`);
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }
  if (!isValidIanaTimezoneShape(timezone)) {
    console.warn(`[settings] action rejected: timezone shape invalid (length=${timezone.length})`);
    throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
  }

  const goalResult = await savePreferenceField("dailyGoal", () =>
    updateUserDailyGoal(request, accessToken, dailyGoal),
  );
  const tzResult = await savePreferenceField("timezone", () =>
    updateUserTimezone(request, accessToken, timezone),
  );

  const failedFields = [goalResult, tzResult].filter((r) => !r.ok).map((r) => r.field);
  if (failedFields.length > 0) {
    return { ok: false as const, intent: "saveStudyPreferences" as const, failedFields };
  }
  return {
    ok: true as const,
    intent: "saveStudyPreferences" as const,
    savedAt: new Date().toISOString(),
  };
}

export async function action({ request }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const formData = await request.formData();
  const intent = String(formData.get("intent") ?? "");

  if (intent === "saveLanguage") {
    return handleLanguageIntent(request, accessToken, formData);
  }
  if (intent === "saveStudyPreferences") {
    return handleStudyIntent(request, accessToken, formData);
  }
  // Never interpolate the raw intent — a tampered form body could pack
  // an injection payload or control characters that poison downstream
  // log viewers. Length is enough to distinguish "empty submit" from
  // "garbage submit" while staying compliance-safe.
  console.warn(`[settings] action rejected: unknown intent (length=${intent.length})`);
  throw new Response(REJECTED_REQUEST_BODY, { status: 400 });
}

// Discriminator for the language fetcher's action payload — narrowed
// inline so a tampered/legacy fetcher.data shape (no `intent` field, or
// a study-intent shape) is rejected before we try to commit it.
type LanguageActionData = {
  ok: true;
  intent: "saveLanguage";
  language: string;
};

function isLanguageSuccess(data: unknown): data is LanguageActionData {
  if (data === null || typeof data !== "object") return false;
  const obj = data as Record<string, unknown>;
  return (
    obj.ok === true &&
    obj.intent === "saveLanguage" &&
    typeof obj.language === "string" &&
    supportedLanguages.includes(obj.language as SupportedLanguage)
  );
}

export default function SettingsPage() {
  // `language` from the loader is intentionally not read: i18next is the
  // runtime source of truth on the client (driven by cookie/localStorage),
  // and we only re-persist it on user-initiated changes via the language
  // fetcher. The loader still returns it so future surfaces (and tests)
  // can verify the round-trip without a second /auth/me call.
  const { loginId, organizationName, dailyGoal, timezone } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const { t, i18n } = useTranslation();
  const languageFetcher = useFetcher<typeof action>();

  // Draft state is seeded once and stays under the user's control. We
  // intentionally do not re-sync from loader on every revalidation so
  // partial-failure retries do not wipe mid-edit input.
  const [draftGoal, setDraftGoal] = useState<number>(dailyGoal);
  const [draftTimezone, setDraftTimezone] = useState(timezone);

  // The language the user has *requested* but the server has not yet
  // confirmed. Drives the controlled select so the picker shows the
  // pending value while disabled; collapses to i18n.language when no
  // request is in flight.
  const [pendingLanguage, setPendingLanguage] = useState<SupportedLanguage | null>(null);

  const isSavingStudy =
    navigation.state === "submitting" &&
    navigation.formData?.get("intent") === "saveStudyPreferences";
  const isChangingLanguage = languageFetcher.state !== "idle";

  // Browser-detected timezone never changes during a mount; memoise to
  // avoid recomputing Intl.DateTimeFormat on each render.
  const detectedBrowserTimezone = useMemo(() => detectBrowserTimezone(), []);

  // Commit i18n.changeLanguage only after the server confirms the write.
  // Without this gate a network/4xx/5xx failure would leave the UI on
  // the new language while the persisted preference stayed on the old
  // one — a silent inconsistency that survives until the user manually
  // re-changes. Tracking the last data payload we acted on via a ref
  // prevents the effect from re-firing on every render once it has
  // committed (fetcher.data is stable until the next submit, but the
  // useEffect dep list cannot tell us that).
  const lastCommittedDataRef = useRef<unknown>(null);
  useEffect(() => {
    if (languageFetcher.state !== "idle") return;
    const data = languageFetcher.data;
    if (data === undefined || data === lastCommittedDataRef.current) return;
    lastCommittedDataRef.current = data;

    if (isLanguageSuccess(data)) {
      if (i18n.language !== data.language) {
        i18n.changeLanguage(data.language);
      }
      setPendingLanguage(null);
      return;
    }
    // Server rejected the language save (validation failure, network
    // error, or any future non-ok payload). Drop the pending value so
    // the select snaps back to the still-persisted i18n.language.
    setPendingLanguage(null);
  }, [languageFetcher.state, languageFetcher.data, i18n]);

  function handleLanguageChange(event: React.ChangeEvent<HTMLSelectElement>) {
    const nextLanguage = event.target.value as SupportedLanguage;
    if (nextLanguage === i18n.language) return;
    // Show the user's choice in the picker immediately, but do NOT call
    // i18n.changeLanguage until the server has accepted the write. See
    // the effect above for the commit path.
    setPendingLanguage(nextLanguage);
    languageFetcher.submit({ intent: "saveLanguage", language: nextLanguage }, { method: "post" });
  }

  const studySaved = actionData?.ok === true && actionData.intent === "saveStudyPreferences";
  const studyPartialFailure =
    actionData?.ok === false && actionData.intent === "saveStudyPreferences";
  const selectedLanguage = pendingLanguage ?? (i18n.language as SupportedLanguage);

  return (
    <div className="space-y-6">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold">{t("settings.title")}</h1>
        <p className="text-sm text-muted-foreground">{t("settings.subtitle")}</p>
      </header>

      <section
        aria-labelledby="settings-account-heading"
        className="space-y-3 rounded-md border bg-card p-4"
      >
        <h2 id="settings-account-heading" className="text-sm font-semibold">
          {t("settings.account.heading")}
        </h2>
        <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-2 text-sm">
          <dt className="text-muted-foreground">{t("settings.account.loginIdLabel")}</dt>
          <dd className="break-all" data-testid="settings-login-id">
            {loginId}
          </dd>
          <dt className="text-muted-foreground">{t("settings.account.organizationLabel")}</dt>
          <dd className="break-all" data-testid="settings-organization">
            {organizationName}
          </dd>
        </dl>
      </section>

      <section
        aria-labelledby="settings-language-heading"
        className="space-y-3 rounded-md border bg-card p-4"
      >
        <h2 id="settings-language-heading" className="text-sm font-semibold">
          {t("settings.language.heading")}
        </h2>
        <div>
          <label htmlFor="language" className="block text-sm font-medium">
            {t("settings.language.label")}
          </label>
          <select
            id="language"
            value={selectedLanguage}
            onChange={handleLanguageChange}
            disabled={isChangingLanguage}
            aria-busy={isChangingLanguage}
            data-testid="settings-language-select"
            className="mt-1 h-9 w-48 rounded-md border border-input bg-transparent px-2 text-sm"
          >
            {supportedLanguages.map((lang) => (
              <option key={lang} value={lang}>
                {t(`languages.${lang}`, { defaultValue: lang.toUpperCase() })}
              </option>
            ))}
          </select>
          <p className="mt-1 text-xs text-muted-foreground">{t("settings.language.hint")}</p>
        </div>
      </section>

      <section
        aria-labelledby="settings-study-heading"
        className="space-y-4 rounded-md border bg-card p-4"
      >
        <h2 id="settings-study-heading" className="text-sm font-semibold">
          {t("settings.study.heading")}
        </h2>
        <Form method="post" className="space-y-3">
          <input type="hidden" name="intent" value="saveStudyPreferences" />
          <div>
            <label htmlFor="dailyGoal" className="block text-sm font-medium">
              {t("settings.study.dailyGoalLabel")}
            </label>
            {/*
              Treating empty as NaN (instead of Number("") === 0) keeps the
              field visually empty when the user clears it and prevents
              the "0" submission the server would reject. `required` blocks
              submit while empty so the user gets browser-native feedback.
            */}
            <input
              id="dailyGoal"
              name="dailyGoal"
              type="number"
              min={MIN_DAILY_GOAL}
              max={MAX_DAILY_GOAL}
              required
              value={Number.isFinite(draftGoal) ? draftGoal : ""}
              onChange={(e) => {
                const raw = e.target.value;
                setDraftGoal(raw === "" ? Number.NaN : Number(raw));
              }}
              className="mt-1 h-9 w-32 rounded-md border border-input bg-transparent px-2 text-sm"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              {t("settings.study.dailyGoalHint")}
            </p>
          </div>
          <div>
            <label htmlFor="timezone" className="block text-sm font-medium">
              {t("settings.study.timezoneLabel")}
            </label>
            <input
              id="timezone"
              name="timezone"
              type="text"
              value={draftTimezone}
              onChange={(e) => setDraftTimezone(e.target.value)}
              maxLength={64}
              pattern="[A-Za-z_/+\-0-9]+"
              className="mt-1 h-9 w-72 rounded-md border border-input bg-transparent px-2 text-sm"
            />
            <p className="mt-1 text-xs text-muted-foreground">
              {t("settings.study.timezoneHint")}
              <br />
              {t("settings.study.timezoneDetected", { timezone: detectedBrowserTimezone })}
            </p>
          </div>
          <Button type="submit" disabled={isSavingStudy}>
            {isSavingStudy ? t("common.saving") : t("settings.study.save")}
          </Button>
          {studySaved && (
            <p className="text-xs text-emerald-600" data-testid="settings-saved">
              {t("settings.study.saved")}
            </p>
          )}
          {studyPartialFailure && (
            <p className="text-xs text-amber-700" data-testid="settings-partial-failure">
              {t("settings.study.partialFailure", {
                fields: actionData.failedFields.join(", "),
              })}
            </p>
          )}
        </Form>
      </section>
    </div>
  );
}
