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
    return { error: "Email and password are required." };
  }

  const headers = new Headers();
  const supabase = createSupabaseServerClient(request, headers);

  const { data, error } = await supabase.auth.signUp({ email, password });
  if (error) {
    console.error("signup: signUp failed", error);
    return { error: error.message };
  }
  if (!data.session) {
    return { message: "Check your email to confirm your account." };
  }

  const organizationName = process.env.ORGANIZATION_NAME ?? "";
  let tokens: Awaited<ReturnType<typeof exchangeSupabaseToken>>;
  try {
    tokens = await exchangeSupabaseToken(data.session.access_token, organizationName);
  } catch (e) {
    console.error("signup: exchangeSupabaseToken failed", e);
    return { error: "Authentication service is temporarily unavailable." };
  }

  const session = await getSession(request);
  session.set("accessToken", tokens.accessToken);
  session.set("refreshToken", tokens.refreshToken);
  headers.append("Set-Cookie", await commitSession(session));

  return redirect("/", { headers });
}

export default function SignupPage() {
  const actionData = useActionData<typeof action>();

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6 p-6">
        <h1 className="text-center text-2xl font-bold">Sign Up</h1>

        {actionData?.error && (
          <p className="text-center text-sm text-red-600">{actionData.error}</p>
        )}
        {actionData && "message" in actionData && (
          <p className="text-center text-sm text-green-600">{actionData.message}</p>
        )}

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
              autoComplete="new-password"
              minLength={8}
            />
          </div>
          <Button type="submit" className="w-full">
            Sign Up
          </Button>
        </Form>

        <p className="text-center text-sm">
          Already have an account?{" "}
          <a href="/login" className="underline">
            Login
          </a>
        </p>
      </div>
    </div>
  );
}
