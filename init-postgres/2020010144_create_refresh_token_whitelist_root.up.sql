create table refresh_token_whitelist_root (
  user_id uuid not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,primary key(user_id)
 ,foreign key(user_id) references app_user(id) on delete cascade
);

CREATE TRIGGER update_refresh_token_whitelist_root_updated_at
    BEFORE UPDATE ON refresh_token_whitelist_root
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
