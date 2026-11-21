create table user_n_group_root (
  group_id uuid not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,primary key(group_id)
 ,foreign key(group_id) references "group"(id) on delete cascade
);

CREATE TRIGGER update_user_n_group_root_updated_at
    BEFORE UPDATE ON user_n_group_root
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
