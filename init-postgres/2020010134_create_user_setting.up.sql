create table user_setting (
 app_user_id uuid not null
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by uuid not null
,updated_by uuid not null
,max_workbooks int not null default 3
,primary key(app_user_id)
,foreign key(app_user_id) references app_user(id) on delete cascade
);

CREATE TRIGGER update_user_setting_updated_at
    BEFORE UPDATE ON user_setting
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
