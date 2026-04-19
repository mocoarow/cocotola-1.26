-- RLS: refresh_token_whitelist
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.refresh_token_whitelist ENABLE ROW LEVEL SECURITY;

CREATE POLICY "refresh_token_whitelist_select"
  ON public.refresh_token_whitelist FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());
