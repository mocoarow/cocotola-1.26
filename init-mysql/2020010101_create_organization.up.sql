create table `organization` (
 `id` char(36) character set ascii not null
,`version` int not null default 1
,`created_at` datetime not null default current_timestamp
,`updated_at` datetime not null default current_timestamp on update current_timestamp
,`created_by` char(36) character set ascii not null
,`updated_by` char(36) character set ascii not null
,`name` varchar(255) character set ascii not null
,`max_active_users` int not null default 100
,`max_active_groups` int not null default 100
,primary key(`id`)
,unique(`name`)
);
