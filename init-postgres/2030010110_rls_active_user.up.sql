-- RLS: active_user
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.active_user ENABLE ROW LEVEL SECURITY;

CREATE POLICY "active_user_select"
  ON public.active_user FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
