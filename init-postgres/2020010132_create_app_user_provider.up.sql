create table app_user_provider (
 id uuid not null
,version int not null default 1
,created_at timestamp not null default current_timestamp
,updated_at timestamp not null default current_timestamp
,created_by uuid not null
,updated_by uuid not null
,app_user_id uuid not null
,organization_id uuid not null
,provider varchar(40) not null
,provider_id varchar(200) not null
,primary key(id)
,unique(organization_id, provider, provider_id)
,foreign key(app_user_id) references app_user(id) on delete cascade
,foreign key(organization_id) references organization(id) on delete cascade
);

CREATE TRIGGER update_app_user_provider_updated_at
    BEFORE UPDATE ON app_user_provider
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
