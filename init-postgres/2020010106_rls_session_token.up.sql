-- RLS: session_token
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.session_token ENABLE ROW LEVEL SECURITY;

CREATE POLICY "session_token_select"
  ON public.session_token FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());

CREATE POLICY "session_token_delete"
  ON public.session_token FOR DELETE
  TO authenticated
  USING (user_id = public.current_app_user_id());
