import { redirect } from "react-router";
import { destroySession, getSession } from "~/lib/auth/session.server";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import type { Route } from "./+types/logout";

export async function action({ request }: Route.ActionArgs) {
  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);
  await supabase.auth.signOut();

  const session = await getSession(request);
  headers.append("Set-Cookie", await destroySession(session));

  return redirect("/login", { headers });
}

export async function loader(_args: Route.LoaderArgs) {
  return redirect("/");
}
