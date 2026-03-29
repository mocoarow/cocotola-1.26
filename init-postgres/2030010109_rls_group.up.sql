-- RLS: group
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public."group" ENABLE ROW LEVEL SECURITY;

CREATE POLICY "group_select"
  ON public."group" FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
