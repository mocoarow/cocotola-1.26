import { redirect } from "react-router";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/home";

export function meta(_args: Route.MetaArgs) {
  return [{ title: "Cocotola" }, { name: "description", content: "Cocotola - Learning Platform" }];
}

export async function loader({ context }: Route.LoaderArgs) {
  requireAuth(context);
  throw redirect("/workbooks");
}

export default function Home() {
  return null;
}
