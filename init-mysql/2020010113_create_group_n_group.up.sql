create table `group_n_group` (
 `organization_id` char(36) character set ascii not null
,`parent_group_id` char(36) character set ascii not null
,`child_group_id` char(36) character set ascii not null
,`created_at` datetime not null default current_timestamp
,`created_by` char(36) character set ascii not null
,primary key(`organization_id`, `parent_group_id`, `child_group_id`)
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
,foreign key(`parent_group_id`) references `group`(`id`) on delete cascade
,foreign key(`child_group_id`) references `group`(`id`) on delete cascade
);
