create table user_n_group (
 group_id uuid not null
,user_id uuid not null
,created_at timestamp not null default current_timestamp
,created_by uuid not null
,primary key(group_id, user_id)
,foreign key(group_id) references "group"(id) on delete cascade
,foreign key(user_id) references app_user(id) on delete cascade
);
