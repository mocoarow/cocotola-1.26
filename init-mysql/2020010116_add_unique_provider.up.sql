ALTER TABLE `app_user`
  ADD UNIQUE INDEX `uq_app_user_provider` (`organization_id`, `provider`, `provider_id`);
