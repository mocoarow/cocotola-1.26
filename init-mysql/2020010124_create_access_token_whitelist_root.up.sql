create table `access_token_whitelist_root` (
  `user_id` char(36) character set ascii not null
 ,`version` int not null default 1
 ,`created_at` datetime not null default current_timestamp
 ,`updated_at` datetime not null default current_timestamp on update current_timestamp
 ,primary key(`user_id`)
 ,foreign key(`user_id`) references `app_user`(`id`) on delete cascade
);
