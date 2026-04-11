CREATE TABLE `app_user_provider` (
  `id` CHAR(36) NOT NULL,
  `version` INT NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `created_by` CHAR(36) NOT NULL,
  `updated_by` CHAR(36) NOT NULL,
  `app_user_id` CHAR(36) NOT NULL,
  `organization_id` CHAR(36) NOT NULL,
  `provider` VARCHAR(40) NOT NULL,
  `provider_id` VARCHAR(200) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_app_user_provider_org_provider` (`organization_id`, `provider`, `provider_id`),
  CONSTRAINT `fk_app_user_provider_user` FOREIGN KEY (`app_user_id`) REFERENCES `app_user` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_app_user_provider_org` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON DELETE CASCADE
);
