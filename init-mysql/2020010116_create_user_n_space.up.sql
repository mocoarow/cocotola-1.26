create table `user_n_space` (
 `id` int auto_increment
,`created_at` datetime not null default current_timestamp
,`created_by` int not null
,`organization_id` int not null
,`user_id` int not null
,`space_id` int not null
,primary key(`id`)
,unique(`organization_id`, `user_id`, `space_id`)
,foreign key(`created_by`) references `app_user`(`id`) on delete cascade
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
,foreign key(`space_id`) references `space`(`id`) on delete cascade
);
