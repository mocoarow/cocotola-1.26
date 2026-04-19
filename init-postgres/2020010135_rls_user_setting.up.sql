-- RLS: user_setting
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.user_setting ENABLE ROW LEVEL SECURITY;

CREATE POLICY "user_setting_select"
  ON public.user_setting FOR SELECT
  TO authenticated
  USING (app_user_id = public.current_app_user_id());
