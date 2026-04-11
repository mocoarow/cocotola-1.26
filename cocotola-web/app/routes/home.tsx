import { redirect } from "react-router";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/home";

export function meta(_args: Route.MetaArgs) {
  return [{ title: "Cocotola" }, { name: "description", content: "Cocotola - Learning Platform" }];
}

export async function loader({ request }: Route.LoaderArgs) {
  await requireAuth(request);
  throw redirect("/workbooks");
}

export default function Home() {
  return null;
}
