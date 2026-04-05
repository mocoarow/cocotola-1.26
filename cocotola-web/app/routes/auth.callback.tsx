import { redirect } from "react-router";
import { exchangeSupabaseToken } from "~/lib/api/auth.server";
import { commitSession, getSession } from "~/lib/auth/session.server";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import type { Route } from "./+types/auth.callback";

export async function loader({ request }: Route.LoaderArgs) {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");

  if (!code) {
    return redirect("/login");
  }

  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);

  const { data, error } = await supabase.auth.exchangeCodeForSession(code);
  if (error || !data.session) {
    console.error("auth.callback: exchangeCodeForSession failed", error);
    return redirect("/login");
  }

  const organizationName = process.env.ORGANIZATION_NAME ?? "";
  let tokens: Awaited<ReturnType<typeof exchangeSupabaseToken>>;
  try {
    tokens = await exchangeSupabaseToken(data.session.access_token, organizationName);
  } catch (e) {
    console.error("auth.callback: exchangeSupabaseToken failed", e);
    return redirect("/login");
  }

  const session = await getSession(request);
  session.set("accessToken", tokens.accessToken);
  session.set("refreshToken", tokens.refreshToken);
  headers.append("Set-Cookie", await commitSession(session));

  return redirect("/", { headers });
}

export default function AuthCallback() {
  return <p>Authenticating...</p>;
}
