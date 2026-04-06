import { Form, redirect, useActionData } from "react-router";
import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { exchangeSupabaseToken } from "~/lib/api/auth.server";
import { commitSession, getSession } from "~/lib/auth/session.server";
import { createSupabaseBrowserClient } from "~/lib/supabase/client";
import { createSupabaseServerClient } from "~/lib/supabase/server";
import type { Route } from "./+types/login";

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
    return { error: "Email and password are required." };
  }

  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);

  const { data, error } = await supabase.auth.signInWithPassword({ email, password });
  if (error || !data.session) {
    console.error("login: signInWithPassword failed", error);
    return { error: error?.message ?? "Authentication failed." };
  }

  const organizationName = process.env.ORGANIZATION_NAME ?? "";
  let tokens: Awaited<ReturnType<typeof exchangeSupabaseToken>>;
  try {
    tokens = await exchangeSupabaseToken(data.session.access_token, organizationName);
  } catch (e) {
    console.error("login: exchangeSupabaseToken failed", e);
    return { error: "Authentication service is temporarily unavailable." };
  }

  const session = await getSession(request);
  session.set("accessToken", tokens.accessToken);
  session.set("refreshToken", tokens.refreshToken);
  headers.append("Set-Cookie", await commitSession(session));

  return redirect("/", { headers });
}

export default function LoginPage() {
  const actionData = useActionData<typeof action>();

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
        <h1 className="text-center text-2xl font-bold">Login</h1>

        {actionData?.error && (
          <p className="text-center text-sm text-red-600">{actionData.error}</p>
        )}

        <Button type="button" variant="outline" className="w-full" onClick={handleGoogleLogin}>
          Sign in with Google
        </Button>

        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-background px-2 text-muted-foreground">or</span>
          </div>
        </div>

        <Form method="post" className="space-y-4">
          <div>
            <label htmlFor="email" className="block text-sm font-medium">
              Email
            </label>
            <Input id="email" name="email" type="email" required autoComplete="email" />
          </div>
          <div>
            <label htmlFor="password" className="block text-sm font-medium">
              Password
            </label>
            <Input
              id="password"
              name="password"
              type="password"
              required
              autoComplete="current-password"
              minLength={8}
            />
          </div>
          <Button type="submit" className="w-full">
            Login
          </Button>
        </Form>

        <p className="text-center text-sm">
          Don&apos;t have an account?{" "}
          <a href="/signup" className="underline">
            Sign up
          </a>
        </p>
      </div>
    </div>
  );
}
