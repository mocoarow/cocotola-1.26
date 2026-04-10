create table space (
 id uuid not null default gen_random_uuid()
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by uuid not null
,updated_by uuid not null
,organization_id uuid not null
,owner_id uuid not null
,key_name varchar(50) not null
,name varchar(100) not null
,space_type varchar(20) not null
,deleted boolean not null default false
,primary key(id)
,unique(organization_id, key_name)
,foreign key(created_by) references app_user(id) on delete cascade
,foreign key(updated_by) references app_user(id) on delete cascade
,foreign key(organization_id) references organization(id) on delete cascade
,foreign key(owner_id) references app_user(id) on delete cascade
);

CREATE TRIGGER update_space_updated_at
    BEFORE UPDATE ON space
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
