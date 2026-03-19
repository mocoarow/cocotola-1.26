create table `session_token` (
  `id` varchar(36) character set ascii not null
 ,`version` int not null default 1
 ,`created_at` datetime not null default current_timestamp
 ,`updated_at` datetime not null default current_timestamp on update current_timestamp
 ,`user_id` int not null
 ,`login_id` varchar(200) character set ascii not null
 ,`organization_name` varchar(20) character set ascii not null
 ,`token_hash` varchar(64) character set ascii not null
 ,`expires_at` datetime not null
 ,`revoked_at` datetime
 ,primary key(`id`)
 ,unique(`token_hash`)
 ,index(`user_id`, `created_at`)
 ,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
);
