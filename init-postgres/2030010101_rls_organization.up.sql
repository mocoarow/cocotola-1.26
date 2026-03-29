-- RLS: organization
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.organization ENABLE ROW LEVEL SECURITY;

CREATE POLICY "organization_select"
  ON public.organization FOR SELECT
  TO authenticated
  USING (id = public.current_organization_id());
