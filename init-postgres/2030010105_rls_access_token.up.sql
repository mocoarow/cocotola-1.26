-- RLS: access_token
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.access_token ENABLE ROW LEVEL SECURITY;

CREATE POLICY "access_token_select"
  ON public.access_token FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());

CREATE POLICY "access_token_delete"
  ON public.access_token FOR DELETE
  TO authenticated
  USING (user_id = public.current_app_user_id());
