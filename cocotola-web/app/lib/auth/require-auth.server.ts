import { redirect } from "react-router";
import { getSession } from "./session.server";

export async function requireAuth(request: Request) {
  const session = await getSession(request);
  const accessToken = session.get("accessToken");

  if (!accessToken) {
    throw redirect("/login");
  }

  return { accessToken, refreshToken: session.get("refreshToken") };
}
