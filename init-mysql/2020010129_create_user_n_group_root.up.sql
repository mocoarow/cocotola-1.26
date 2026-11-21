create table `user_n_group_root` (
  `group_id` char(36) character set ascii not null
 ,`version` int not null default 1
 ,`created_at` datetime not null default current_timestamp
 ,`updated_at` datetime not null default current_timestamp on update current_timestamp
 ,primary key(`group_id`)
 ,foreign key(`group_id`) references `group`(`id`) on delete cascade
);
