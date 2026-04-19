-- RLS: space
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.space ENABLE ROW LEVEL SECURITY;

CREATE POLICY "space_select"
  ON public.space FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());

CREATE POLICY "space_insert"
  ON public.space FOR INSERT
  TO authenticated
  WITH CHECK (
    organization_id = public.current_organization_id()
    AND owner_id = public.current_app_user_id()
  );

CREATE POLICY "space_update"
  ON public.space FOR UPDATE
  TO authenticated
  USING (owner_id = public.current_app_user_id())
  WITH CHECK (owner_id = public.current_app_user_id());

CREATE POLICY "space_delete"
  ON public.space FOR DELETE
  TO authenticated
  USING (owner_id = public.current_app_user_id());
