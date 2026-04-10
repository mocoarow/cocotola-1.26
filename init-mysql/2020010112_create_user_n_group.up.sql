create table `user_n_group` (
 `group_id` char(36) character set ascii not null
,`user_id` char(36) character set ascii not null
,`created_at` datetime not null default current_timestamp
,`created_by` char(36) character set ascii not null
,primary key(`group_id`, `user_id`)
,foreign key(`group_id`) references `group`(`id`) on delete cascade
,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
);
