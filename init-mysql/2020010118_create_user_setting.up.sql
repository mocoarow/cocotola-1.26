create table `user_setting` (
 `app_user_id` char(36) character set ascii not null
,`version` int not null default 1
,`created_at` datetime not null default current_timestamp
,`updated_at` datetime not null default current_timestamp on update current_timestamp
,`created_by` char(36) character set ascii not null
,`updated_by` char(36) character set ascii not null
,`max_workbooks` int not null
,primary key(`app_user_id`)
,foreign key(`app_user_id`) references `app_user`(`id`) on delete cascade
);
