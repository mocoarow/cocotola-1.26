-- RLS: active_group
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.active_group ENABLE ROW LEVEL SECURITY;

CREATE POLICY "active_group_select"
  ON public.active_group FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
