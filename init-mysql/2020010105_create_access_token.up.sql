create table `access_token` (
  `id` varchar(36) character set ascii not null
 ,`version` int not null default 1
 ,`created_at` datetime not null default current_timestamp
 ,`updated_at` datetime not null default current_timestamp on update current_timestamp
 ,`refresh_token_id` varchar(36) character set ascii not null
 ,`user_id` int not null
 ,`login_id` varchar(200) character set ascii not null
 ,`organization_name` varchar(20) character set ascii not null
 ,`expires_at` datetime not null
 ,`revoked_at` datetime
 ,primary key(`id`)
 ,index(`user_id`, `created_at`)
 ,index(`refresh_token_id`)
 ,foreign key(`user_id`) references `user`(`id`) on delete cascade
 ,foreign key(`refresh_token_id`) references `refresh_token`(`id`) on delete cascade
);
