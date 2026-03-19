create table `access_token_whitelist` (
  `user_id` int not null
 ,`token_id` varchar(36) character set ascii not null
 ,`created_at` datetime not null default current_timestamp
 ,primary key(`user_id`, `token_id`)
 ,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
 ,foreign key(`token_id`) references `access_token`(`id`) on delete cascade
);
