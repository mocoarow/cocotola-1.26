-- RLS: healthcheck2
-- No policy for authenticated = deny all direct access (managed by backend only)
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.healthcheck2 ENABLE ROW LEVEL SECURITY;
