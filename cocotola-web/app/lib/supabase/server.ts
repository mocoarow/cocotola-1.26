import { createServerClient, parseCookieHeader, serializeCookieHeader } from "@supabase/ssr";

export function createSupabaseServerClient(request: Request, headers: Headers) {
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
  if (!supabaseUrl) {
    throw new Error("VITE_SUPABASE_URL environment variable is required");
  }

  const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY;
  if (!supabaseAnonKey) {
    throw new Error("VITE_SUPABASE_ANON_KEY environment variable is required");
  }

  return createServerClient(supabaseUrl, supabaseAnonKey, {
    cookies: {
      getAll() {
        return parseCookieHeader(request.headers.get("Cookie") ?? "").filter(
          (c): c is { name: string; value: string } => c.value !== undefined,
        );
      },
      setAll(cookiesToSet) {
        for (const { name, value, options } of cookiesToSet) {
          headers.append("Set-Cookie", serializeCookieHeader(name, value, options));
        }
      },
    },
  });
}
