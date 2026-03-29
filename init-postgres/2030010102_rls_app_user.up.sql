-- RLS: app_user
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.app_user ENABLE ROW LEVEL SECURITY;

CREATE POLICY "app_user_select"
  ON public.app_user FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
