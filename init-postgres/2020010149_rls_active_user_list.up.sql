-- RLS: active_user_list
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.active_user_list ENABLE ROW LEVEL SECURITY;

CREATE POLICY "active_user_list_select"
  ON public.active_user_list FOR SELECT
  TO authenticated
  USING (organization_id = public.current_organization_id());
