-- RLS: casbin_rule
-- No policy for authenticated = deny all direct access (managed by backend only)
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.casbin_rule ENABLE ROW LEVEL SECURITY;
