const METADATA_URL =
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity";

function isLocalOrTest(): boolean {
  const appEnv = process.env.APP_ENV ?? "local";
  return appEnv === "local" || appEnv === "test";
}

export async function getIdToken(audience: string): Promise<string | undefined> {
  console.info(
    `[id-token] getIdToken called: audience=${audience}, APP_ENV=${process.env.APP_ENV}`,
  );

  if (isLocalOrTest()) {
    console.info("[id-token] skipping ID token (local or test environment)");
    return undefined;
  }

  try {
    const url = `${METADATA_URL}?audience=${encodeURIComponent(audience)}`;
    console.info(`[id-token] fetching ID token from metadata server: url=${url}`);

    const response = await fetch(url, {
      headers: { "Metadata-Flavor": "Google" },
    });

    if (!response.ok) {
      const body = await response.text();
      console.error(
        `[id-token] metadata server returned error: status=${response.status}, body=${body}`,
      );
      throw new Error(
        `Failed to obtain ID token for audience: ${audience} (status=${response.status})`,
      );
    }

    const token = await response.text();
    console.info(
      `[id-token] ID token obtained successfully: length=${token.length}, prefix=${token.substring(0, 20)}...`,
    );
    return token;
  } catch (error) {
    console.error(`[id-token] failed to obtain ID token for audience=${audience}:`, error);
    throw error;
  }
}
