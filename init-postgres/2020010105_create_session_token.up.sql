create table session_token (
  id varchar(36) not null
 ,version int not null default 1
 ,created_at timestamp not null default current_timestamp
 ,updated_at timestamp not null default current_timestamp
 ,user_id uuid not null
 ,login_id varchar(200) not null
 ,organization_name varchar(20) not null
 ,token_hash varchar(64) not null
 ,expires_at timestamp not null
 ,revoked_at timestamp
 ,primary key(id)
 ,unique(token_hash)
 ,foreign key(user_id) references app_user(id) on delete cascade
);

CREATE INDEX idx_session_token_user_id_created_at ON session_token(user_id, created_at);

CREATE TRIGGER update_session_token_updated_at
    BEFORE UPDATE ON session_token
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
