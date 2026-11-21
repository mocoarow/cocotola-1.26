create table active_user_list (
  organization_id uuid not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,primary key(organization_id)
 ,foreign key(organization_id) references organization(id) on delete cascade
);

CREATE TRIGGER update_active_user_list_updated_at
    BEFORE UPDATE ON active_user_list
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
