create table `casbin_rule` (
 `id` bigint unsigned auto_increment
,`ptype` varchar(100) not null default ''
,`v0` varchar(100) not null default ''
,`v1` varchar(100) not null default ''
,`v2` varchar(100) not null default ''
,`v3` varchar(100) not null default ''
,`v4` varchar(100) not null default ''
,`v5` varchar(100) not null default ''
,primary key(`id`)
,unique key `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`)
);
