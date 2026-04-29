import { redirect } from "react-router";
import { type SupportedLanguage, supportedLanguages } from "~/i18n/config";
import { updateUserLanguage } from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/user-language";

export async function action({ request }: Route.ActionArgs) {
  const { accessToken } = await requireAuth(request);
  const formData = await request.formData();
  const language = formData.get("language");
  if (typeof language !== "string" || !supportedLanguages.includes(language as SupportedLanguage)) {
    throw new Response("language is invalid", { status: 400 });
  }
  await updateUserLanguage(accessToken, language);
  return { ok: true };
}

export async function loader(_args: Route.LoaderArgs) {
  return redirect("/");
}
