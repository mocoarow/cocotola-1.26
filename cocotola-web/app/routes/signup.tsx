import { useTranslation } from "react-i18next";
import { Form, redirect, useActionData } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { exchangeSupabaseToken } from "~/lib/api/auth.server";
import { commitSession, getSession } from "~/lib/auth/session.server";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import type { Route } from "./+types/signup";

export async function loader({ request }: Route.LoaderArgs) {
  const session = await getSession(request);
  if (session.get("accessToken")) {
    return redirect("/");
  }
  return null;
}

export async function action({ request }: Route.ActionArgs) {
  const formData = await request.formData();
  const email = String(formData.get("email") ?? "");
  const password = String(formData.get("password") ?? "");

  if (!email || !password) {
    return { errorKey: "signup.errors.emailPasswordRequired" };
  }

  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);

  const { data, error } = await supabase.auth.signUp({ email, password });
  if (error) {
    console.error("signup: signUp failed", error);
    return { errorMessage: error.message };
  }
  if (!data.session) {
    return { messageKey: "signup.confirmEmail" };
  }

  const organizationName = process.env.ORGANIZATION_NAME ?? "";
  let tokens: Awaited<ReturnType<typeof exchangeSupabaseToken>>;
  try {
    tokens = await exchangeSupabaseToken(data.session.access_token, organizationName);
  } catch (e) {
    console.error("signup: exchangeSupabaseToken failed", e);
    return { errorKey: "signup.errors.authUnavailable" };
  }

  const session = await getSession(request);
  session.set("accessToken", tokens.accessToken);
  session.set("refreshToken", tokens.refreshToken);
  headers.append("Set-Cookie", await commitSession(session));

  return redirect("/", { headers });
}

export default function SignupPage() {
  const actionData = useActionData<typeof action>();
  const { t } = useTranslation();

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 p-6">
        <h1 className="text-center text-2xl font-bold">{t("signup.title")}</h1>

        {actionData && "errorKey" in actionData && actionData.errorKey && (
          <p className="text-center text-sm text-red-600">{t(actionData.errorKey)}</p>
        )}
        {actionData && "errorMessage" in actionData && actionData.errorMessage && (
          <p className="text-center text-sm text-red-600">{actionData.errorMessage}</p>
        )}
        {actionData && "messageKey" in actionData && actionData.messageKey && (
          <p className="text-center text-sm text-green-600">{t(actionData.messageKey)}</p>
        )}

        <Form method="post" className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium">
              {t("signup.emailLabel")}
            </label>
            <Input id="email" name="email" type="email" required autoComplete="email" />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium">
              {t("signup.passwordLabel")}
            </label>
            <Input
              id="password"
              name="password"
              type="password"
              required
              autoComplete="new-password"
              minLength={8}
            />
          </div>
          <Button type="submit" className="w-full">
            {t("signup.submitButton")}
          </Button>
        </Form>

        <p className="text-center text-sm">
          {t("signup.alreadyHaveAccount")}{" "}
          <a href="/login" className="underline">
            {t("signup.loginLink")}
          </a>
        </p>
      </div>
    </div>
  );
}
