create table active_group (
 organization_id uuid not null
,group_id int not null
,created_at timestamp not null default current_timestamp
,primary key(organization_id, group_id)
,foreign key(organization_id) references organization(id) on delete cascade
,foreign key(group_id) references "group"(id) on delete cascade
);
