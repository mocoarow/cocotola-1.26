create table `group` (
 `id` int auto_increment
,`version` int not null default 1
,`created_at` datetime not null default current_timestamp
,`updated_at` datetime not null default current_timestamp on update current_timestamp
,`created_by` int not null
,`updated_by` int not null
,`organization_id` int not null
,`name` varchar(255) not null
,`enabled` tinyint(1) not null default 1
,primary key(`id`)
,unique(`organization_id`, `name`)
,foreign key(`organization_id`) references `organization`(`id`) on delete cascade
);
