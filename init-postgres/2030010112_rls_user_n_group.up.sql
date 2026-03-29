-- RLS: user_n_group
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.user_n_group ENABLE ROW LEVEL SECURITY;

CREATE POLICY "user_n_group_select"
  ON public.user_n_group FOR SELECT
  TO authenticated
  USING (user_id = public.current_app_user_id());
