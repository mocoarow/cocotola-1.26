create table access_token (
  id varchar(36) not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,refresh_token_id varchar(36) not null
 ,user_id int not null
 ,login_id varchar(200) not null
 ,organization_name varchar(20) not null
 ,expires_at timestamp not null
 ,revoked_at timestamp
 ,primary key(id)
 ,foreign key(user_id) references app_user(id) on delete cascade
 ,foreign key(refresh_token_id) references refresh_token(id) on delete cascade
);

CREATE INDEX idx_access_token_user_id_created_at ON access_token(user_id, created_at);
CREATE INDEX idx_access_token_refresh_token_id ON access_token(refresh_token_id);

CREATE TRIGGER update_access_token_updated_at
    BEFORE UPDATE ON access_token
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
