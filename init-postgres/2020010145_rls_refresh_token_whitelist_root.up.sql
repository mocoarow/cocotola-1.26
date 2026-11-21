-- RLS: refresh_token_whitelist_root
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.refresh_token_whitelist_root ENABLE ROW LEVEL SECURITY;

CREATE POLICY "refresh_token_whitelist_root_select"
  ON public.refresh_token_whitelist_root FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());
