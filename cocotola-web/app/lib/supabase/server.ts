import { createServerClient, parseCookieHeader, serializeCookieHeader } from "@supabase/ssr";

export function createSupabaseServerClient(request: Request, headers: Headers) {
  return createServerClient(
    import.meta.env.VITE_SUPABASE_URL,
    import.meta.env.VITE_SUPABASE_ANON_KEY,
    {
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
    },
  );
}
