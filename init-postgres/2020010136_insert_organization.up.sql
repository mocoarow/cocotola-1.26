-- Bootstrap "system" organization.
-- These UUIDs must stay in sync with domain.SystemOrganizationIDString
-- and domain.SystemAppUserIDString in cocotola-auth/domain/ids.go.
insert into organization (id, created_by, updated_by, name) values
('00000000-0000-7000-8000-000000000001', '00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000002', 'system');
