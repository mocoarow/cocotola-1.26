-- RLS: user_n_group_root
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.user_n_group_root ENABLE ROW LEVEL SECURITY;

CREATE POLICY "user_n_group_root_select"
  ON public.user_n_group_root FOR SELECT
  TO authenticated
  USING (group_id IN (SELECT id FROM public."group" WHERE organization_id = public.current_organization_id()));
