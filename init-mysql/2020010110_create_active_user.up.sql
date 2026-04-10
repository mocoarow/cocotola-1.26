create table `active_user` (
 `organization_id` char(36) character set ascii not null
,`user_id` char(36) character set ascii not null
,`created_at` datetime not null default current_timestamp
,primary key(`organization_id`, `user_id`)
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
);
