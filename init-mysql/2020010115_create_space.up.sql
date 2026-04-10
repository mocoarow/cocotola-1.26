create table `space` (
 `id` char(36) character set ascii not null
,`version` int not null default 1
,`created_at` datetime not null default current_timestamp
,`updated_at` datetime not null default current_timestamp on update current_timestamp
,`created_by` char(36) character set ascii not null
,`updated_by` char(36) character set ascii not null
,`organization_id` char(36) character set ascii not null
,`owner_id` char(36) character set ascii not null
,`key_name` varchar(50) character set ascii not null
,`name` varchar(100) not null
,`space_type` varchar(20) character set ascii not null
,`deleted` tinyint(1) not null default 0
,primary key(`id`)
,unique(`organization_id`, `key_name`)
,foreign key(`created_by`) references `app_user`(`id`) on delete cascade
,foreign key(`updated_by`) references `app_user`(`id`) on delete cascade
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
,foreign key(`owner_id`) references `app_user`(`id`) on delete cascade
);
