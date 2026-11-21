create table `group_n_group_root` (
  `organization_id` char(36) character set ascii not null
 ,`version` int not null default 1
 ,`created_at` datetime not null default current_timestamp
 ,`updated_at` datetime not null default current_timestamp on update current_timestamp
 ,primary key(`organization_id`)
 ,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
);
