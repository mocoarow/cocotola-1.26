create table "group" (
 id serial
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by int not null
,updated_by int not null
,organization_id int not null
,name varchar(255) not null
,enabled boolean not null default true
,primary key(id)
,unique(organization_id, name)
,foreign key(organization_id) references organization(id) on delete cascade
);

CREATE TRIGGER update_group_updated_at
    BEFORE UPDATE ON "group"
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
