-- RLS: group_n_group
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.group_n_group ENABLE ROW LEVEL SECURITY;

CREATE POLICY "group_n_group_select"
  ON public.group_n_group FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
