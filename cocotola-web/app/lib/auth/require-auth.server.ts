import { type RouterContextProvider, redirect } from "react-router";
import { authContext } from "./refresh-middleware.server";

export function requireAuth(context: Readonly<RouterContextProvider>): { accessToken: string } {
  const auth = context.get(authContext);
  if (!auth) {
    throw redirect("/login");
  }
  return auth;
}
