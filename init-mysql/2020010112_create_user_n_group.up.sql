create table `user_n_group` (
 `group_id` int not null
,`user_id` int not null
,`created_at` datetime not null default current_timestamp
,`created_by` int not null
,primary key(`group_id`, `user_id`)
,foreign key(`group_id`) references `group`(`id`) on delete cascade
,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
);
