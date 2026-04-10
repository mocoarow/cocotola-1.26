create table refresh_token_whitelist (
  user_id uuid not null
 ,token_id varchar(36) not null
 ,created_at timestamp not null default current_timestamp
 ,primary key(user_id, token_id)
 ,foreign key(user_id) references app_user(id) on delete cascade
 ,foreign key(token_id) references refresh_token(id) on delete cascade
);
