import { GoogleAuth } from "google-auth-library";

const authInstances = new Map<string, GoogleAuth>();

function getGoogleAuth(audience: string): GoogleAuth {
  let auth = authInstances.get(audience);
  if (!auth) {
    auth = new GoogleAuth();
    authInstances.set(audience, auth);
  }
  return auth;
}

function isLocalOrTest(): boolean {
  const appEnv = process.env.APP_ENV ?? "local";
  return appEnv === "local" || appEnv === "test";
}

export async function getIdToken(audience: string): Promise<string | undefined> {
  if (isLocalOrTest()) {
    return undefined;
  }

  const auth = getGoogleAuth(audience);
  const client = await auth.getIdTokenClient(audience);
  const headers = await client.getRequestHeaders();
  const authHeader = headers.get("Authorization");
  if (!authHeader) {
    throw new Error(`Failed to obtain ID token for audience: ${audience}`);
  }

  return authHeader.replace(/^Bearer\s+/, "");
}
