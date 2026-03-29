create table active_user (
 organization_id int not null
,user_id int not null
,created_at timestamp not null default current_timestamp
,primary key(organization_id, user_id)
,foreign key(organization_id) references organization(id) on delete cascade
,foreign key(user_id) references app_user(id) on delete cascade
);
