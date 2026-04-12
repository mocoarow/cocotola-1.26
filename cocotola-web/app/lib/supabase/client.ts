import { createBrowserClient } from "@supabase/ssr";

export function createSupabaseBrowserClient() {
  const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
  if (!supabaseUrl) {
    throw new Error("VITE_SUPABASE_URL environment variable is required");
  }

  const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY;
  if (!supabaseAnonKey) {
    throw new Error("VITE_SUPABASE_ANON_KEY environment variable is required");
  }

  return createBrowserClient(supabaseUrl, supabaseAnonKey);
}
