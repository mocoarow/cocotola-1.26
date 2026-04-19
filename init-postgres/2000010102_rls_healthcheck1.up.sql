-- RLS: healthcheck1
-- No policy for authenticated = deny all direct access (managed by backend only)
-- Note: anon role has no policies (all access denied by default)

ALTER TABLE public.healthcheck1 ENABLE ROW LEVEL SECURITY;
