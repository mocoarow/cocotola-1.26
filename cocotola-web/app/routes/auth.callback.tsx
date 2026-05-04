import { redirect } from "react-router";
import { detectLanguageFromRequest } from "~/i18n/config";
import { exchangeSupabaseToken } from "~/lib/api/auth.server";
import { getUserLanguage, updateUserLanguage } from "~/lib/api/user-setting.server";
import { commitSession, getSession } from "~/lib/auth/session.server";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import type { Route } from "./+types/auth.callback";

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");
  console.info(`[auth.callback] loader called: hasCode=${!!code}`);

  if (!code) {
    console.info("[auth.callback] no code parameter, redirecting to /login");
    return redirect("/login");
  }

  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);

  console.info("[auth.callback] exchanging code for session");
  const { data, error } = await supabase.auth.exchangeCodeForSession(code);
  if (error || !data.session) {
    console.error("[auth.callback] exchangeCodeForSession failed:", error);
    return redirect("/login");
  }
  console.info(
    `[auth.callback] exchangeCodeForSession succeeded: userId=${data.session.user.id}, tokenLength=${data.session.access_token.length}`,
  );

  const organizationName = process.env.ORGANIZATION_NAME ?? "";
  console.info(
    `[auth.callback] calling exchangeSupabaseToken: organizationName=${organizationName}`,
  );
  let tokens: Awaited<ReturnType<typeof exchangeSupabaseToken>>;
  try {
    tokens = await exchangeSupabaseToken(data.session.access_token, organizationName);
  } catch (e) {
    console.error("[auth.callback] exchangeSupabaseToken failed:", e);
    return redirect("/login");
  }
  console.info("[auth.callback] exchangeSupabaseToken succeeded");

  const session = await getSession(request);
  session.set("accessToken", tokens.accessToken);
  session.set("refreshToken", tokens.refreshToken);
  headers.append("Set-Cookie", await commitSession(session));

  // Sync the browser-detected UI language to the user's server-side preference
  // ONLY when the persisted language is still the backend default ("en").
  // Without this, public-workbook listings (filtered by user.language on the
  // backend) would silently default to "en" until the user explicitly changes
  // the language picker, even though the UI already shows their language.
  // We deliberately skip the sync when the user has a non-default language so
  // that an explicit language picker choice is not silently overwritten when
  // the same user logs in from a browser with a different Accept-Language.
  const uiLanguage = detectLanguageFromRequest(request);
  const backendDefaultLanguage = "en";
  try {
    const currentLanguage = await getUserLanguage(tokens.accessToken);
    if (currentLanguage !== backendDefaultLanguage) {
      console.info(
        `[auth.callback] user language already customized (${currentLanguage}), skipping sync`,
      );
    } else if (currentLanguage === uiLanguage) {
      console.info(`[auth.callback] user language already in sync: ${currentLanguage}`);
    } else {
      await updateUserLanguage(tokens.accessToken, uiLanguage);
      console.info(`[auth.callback] synced user language: ${uiLanguage}`);
    }
  } catch (e) {
    console.error("[auth.callback] language sync failed (non-fatal):", e);
  }

  console.info("[auth.callback] session created, redirecting to /");
  return redirect("/", { headers });
}

export default function AuthCallback() {
  return <p>Authenticating...</p>;
}
