import { fetchWithIdToken } from "./fetch.server";

function getAuthUrl(): string {
  const url = process.env.AUTH_BASE_URL;
  if (!url) {
    throw new Error("AUTH_BASE_URL environment variable is required");
  }
  return url;
}

/** Updates the authenticated user's preferred language. */
export async function updateUserLanguage(accessToken: string, language: string): Promise<void> {
  console.info(`[user-setting] updateUserLanguage called: language=${language}`);

  const authUrl = getAuthUrl();
  const url = `${authUrl}/api/v1/auth/user-setting/language`;

  const response = await fetchWithIdToken(authUrl, url, {
    method: "PUT",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ language }),
  });

  if (!response.ok) {
    console.error(
      `[user-setting] updateUserLanguage failed: status=${response.status}, url=${url}`,
    );
    throw new Response("Failed to update user language", { status: response.status });
  }

  console.info("[user-setting] updateUserLanguage succeeded");
}
