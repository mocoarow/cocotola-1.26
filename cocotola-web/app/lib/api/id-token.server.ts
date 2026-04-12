import { GoogleAuth } from "google-auth-library";

const authInstances = new Map<string, GoogleAuth>();

function getGoogleAuth(audience: string): GoogleAuth {
  let auth = authInstances.get(audience);
  if (!auth) {
    console.info(`[id-token] creating new GoogleAuth instance for audience=${audience}`);
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
  console.info(`[id-token] getIdToken called: audience=${audience}, APP_ENV=${process.env.APP_ENV}`);

  if (isLocalOrTest()) {
    console.info("[id-token] skipping ID token (local or test environment)");
    return undefined;
  }

  try {
    const auth = getGoogleAuth(audience);
    console.info(`[id-token] requesting ID token client for audience=${audience}`);
    const client = await auth.getIdTokenClient(audience);

    console.info("[id-token] fetching request headers from ID token client");
    const headers = await client.getRequestHeaders();
    console.info(`[id-token] received headers keys: ${Object.keys(headers).join(", ")}`);

    const authHeader = headers["Authorization"];
    if (!authHeader) {
      console.error(`[id-token] Authorization header missing from response. headers=${JSON.stringify(headers)}`);
      throw new Error(`Failed to obtain ID token for audience: ${audience}`);
    }

    const token = authHeader.replace(/^Bearer\s+/, "");
    console.info(`[id-token] ID token obtained successfully: length=${token.length}, prefix=${token.substring(0, 20)}...`);
    return token;
  } catch (error) {
    console.error(`[id-token] failed to obtain ID token for audience=${audience}:`, error);
    throw error;
  }
}
