create table `app_user` (
 `id` char(36) character set ascii not null
,`version` int not null default 1
,`created_at` datetime not null default current_timestamp
,`updated_at` datetime not null default current_timestamp on update current_timestamp
,`created_by` char(36) character set ascii not null
,`updated_by` char(36) character set ascii not null
,`organization_id` char(36) character set ascii not null
,`login_id` varchar(200) character set ascii not null
,`hashed_password` varchar(200) character set ascii
,`username` varchar(40)
,`provider` varchar(40) character set ascii
,`provider_id` varchar(40) character set ascii
,`encrypted_provider_access_token` text character set ascii
,`encrypted_provider_refresh_token` text character set ascii
,`enabled` tinyint(1) not null default 1
,primary key(`id`)
,unique(`organization_id`, `login_id`)
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
);
