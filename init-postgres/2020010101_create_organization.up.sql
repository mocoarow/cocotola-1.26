CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = current_timestamp;
    RETURN NEW;
END;
$$ language 'plpgsql';

create table organization (
 id uuid not null
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by uuid not null
,updated_by uuid not null
,name varchar(255) not null
,max_active_users int not null default 100
,max_active_groups int not null default 100
,primary key(id)
,unique(name)
);

CREATE TRIGGER update_organization_updated_at
    BEFORE UPDATE ON organization
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
