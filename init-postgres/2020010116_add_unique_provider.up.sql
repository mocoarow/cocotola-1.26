CREATE UNIQUE INDEX uq_app_user_provider ON app_user (organization_id, provider, provider_id)
  WHERE provider IS NOT NULL AND provider_id IS NOT NULL;
