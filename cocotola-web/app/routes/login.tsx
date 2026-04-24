import { useTranslation } from "react-i18next";
import { redirect } from "react-router";
import { Button } from "~/components/ui/button";
import { getSession } from "~/lib/auth/session.server";
import { createSupabaseBrowserClient } from "~/lib/supabase/client";
import type { Route } from "./+types/login";

export async function loader({ request }: Route.LoaderArgs) {
  console.info("[login] loader called");
  const session = await getSession(request);
  if (session.get("accessToken")) {
    console.info("[login] user already authenticated, redirecting to /");
    return redirect("/");
  }
  return null;
}

export default function LoginPage() {
  const { t } = useTranslation();

  async function handleGoogleLogin() {
    const supabase = createSupabaseBrowserClient();
    const { error } = await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });
    if (error) {
      console.error("login: signInWithOAuth failed", error);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 p-6">
        <h1 className="text-center text-2xl font-bold">{t("login.title")}</h1>

        <Button type="button" variant="outline" className="w-full" onClick={handleGoogleLogin}>
          {t("login.signInWithGoogle")}
        </Button>
      </div>
    </div>
  );
}
