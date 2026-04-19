-- RLS: app_user_provider
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.app_user_provider ENABLE ROW LEVEL SECURITY;

CREATE POLICY "app_user_provider_select"
  ON public.app_user_provider FOR SELECT
  TO authenticated
  USING (app_user_id = public.current_app_user_id());
