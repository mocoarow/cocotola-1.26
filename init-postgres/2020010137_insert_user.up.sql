-- Bootstrap "__system_admin" user.
-- These UUIDs must stay in sync with domain.SystemOrganizationIDString
-- and domain.SystemAppUserIDString in cocotola-auth/domain/ids.go.
insert into app_user (id, created_by, updated_by, organization_id, login_id, hashed_password, username, enabled) values
('00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000001', '__system_admin', '!locked', 'Administrator', true);
