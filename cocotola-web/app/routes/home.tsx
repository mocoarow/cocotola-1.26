import { Form, redirect, useLoaderData } from "react-router";
import { Button } from "~/components/ui/button";
import { requireAuth } from "~/lib/auth/require-auth.server";
import { destroySession, getSession } from "~/lib/auth/session.server";
import type { Route } from "./+types/home";

export function meta(_args: Route.MetaArgs) {
  return [{ title: "Cocotola" }, { name: "description", content: "Cocotola - Learning Platform" }];
}

export async function loader({ request }: Route.LoaderArgs) {
  const { accessToken } = await requireAuth(request);

  const authUrl = process.env.COCOTOLA_AUTH_URL;
  if (!authUrl) {
    throw new Error("COCOTOLA_AUTH_URL environment variable is required");
  }

  const response = await fetch(`${authUrl}/api/v1/auth/me`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  });

  if (response.status === 401) {
    const session = await getSession(request);
    throw redirect("/login", { headers: { "Set-Cookie": await destroySession(session) } });
  }
  if (!response.ok) {
    return { user: null };
  }

  const user = (await response.json()) as {
    userId: string;
    loginId: string;
    organizationName: string;
  };
  return { user };
}

export default function Home() {
  const { user } = useLoaderData<typeof loader>();

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="space-y-6 p-6 text-center">
        <h1 className="text-2xl font-bold">Cocotola</h1>

        {user ? (
          <div className="space-y-4">
            <p>
              Logged in as <strong>{user.loginId}</strong>
            </p>
            <p className="text-sm text-gray-500">Organization: {user.organizationName}</p>
            <Form method="post" action="/logout">
              <Button variant="outline">Logout</Button>
            </Form>
          </div>
        ) : (
          <p className="text-sm text-gray-500">Could not load user info.</p>
        )}
      </div>
    </div>
  );
}
