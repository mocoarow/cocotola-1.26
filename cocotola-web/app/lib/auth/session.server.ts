import { createCookieSessionStorage } from "react-router";

type SessionData = {
  accessToken: string;
  refreshToken: string;
};

function getSessionSecret(): string {
  const secret = process.env.SESSION_SECRET;
  if (!secret) {
    throw new Error("SESSION_SECRET environment variable is required");
  }
  return secret;
}

const sessionStorage = createCookieSessionStorage<SessionData>({
  cookie: {
    name: "__cocotola_session",
    httpOnly: true,
    path: "/",
    sameSite: "lax",
    secrets: [getSessionSecret()],
    secure: process.env.NODE_ENV === "production",
    maxAge: 60 * 60 * 24 * 7, // 7 days
  },
});

export async function getSession(request: Request) {
  return sessionStorage.getSession(request.headers.get("Cookie"));
}

export async function commitSession(session: Awaited<ReturnType<typeof getSession>>) {
  return sessionStorage.commitSession(session);
}

export async function destroySession(session: Awaited<ReturnType<typeof getSession>>) {
  return sessionStorage.destroySession(session);
}
