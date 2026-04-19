-- RLS: refresh_token
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.refresh_token ENABLE ROW LEVEL SECURITY;

CREATE POLICY "refresh_token_select"
  ON public.refresh_token FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());

CREATE POLICY "refresh_token_delete"
  ON public.refresh_token FOR DELETE
  TO authenticated
  USING (user_id = public.current_app_user_id());
