create table app_user (
 id serial
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by int not null
,updated_by int not null
,organization_id int not null
,login_id varchar(200) not null
,hashed_password varchar(200)
,username varchar(40)
,provider varchar(40)
,provider_id varchar(40)
,encrypted_provider_access_token text
,encrypted_provider_refresh_token text
,enabled boolean not null default true
,primary key(id)
,unique(organization_id, login_id)
,foreign key(organization_id) references organization(id) on delete cascade
);

CREATE TRIGGER update_app_user_updated_at
    BEFORE UPDATE ON app_user
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
