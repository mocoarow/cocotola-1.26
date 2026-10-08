-- This UUID must stay in sync with domain.SystemAppUserIDString in cocotola-auth/domain/ids.go.
insert into `user_setting` (`app_user_id`, `created_by`, `updated_by`, `max_workbooks`) values
('00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000002', '00000000-0000-7000-8000-000000000002', 100)
on duplicate key update `app_user_id` = `app_user_id`;
