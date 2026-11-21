create table group_n_group_root (
  organization_id uuid not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,primary key(organization_id)
 ,foreign key(organization_id) references organization(id) on delete cascade
);

CREATE TRIGGER update_group_n_group_root_updated_at
    BEFORE UPDATE ON group_n_group_root
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
