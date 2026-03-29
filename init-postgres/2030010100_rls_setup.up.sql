-- RLS setup: roles and helper functions
-- Compatible with both Supabase and Docker PostgreSQL

-- Create roles if they don't exist (for Docker compatibility)
-- On Supabase these roles already exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'anon') THEN
    CREATE ROLE anon NOLOGIN;
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'authenticated') THEN
    CREATE ROLE authenticated NOLOGIN;
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'service_role') THEN
    CREATE ROLE service_role NOLOGIN;
  END IF;
END
$$;

-- Helper functions to get current user context from session settings
-- Usage: SET app.current_user_id = '123'; SET app.current_organization_id = '1';
CREATE OR REPLACE FUNCTION public.current_app_user_id() RETURNS int
LANGUAGE sql STABLE
AS $$
  SELECT COALESCE(NULLIF(current_setting('app.current_user_id', true), '')::int, -1);
$$;

CREATE OR REPLACE FUNCTION public.current_organization_id() RETURNS int
LANGUAGE sql STABLE
AS $$
  SELECT COALESCE(NULLIF(current_setting('app.current_organization_id', true), '')::int, -1);
$$;
