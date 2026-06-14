alter table user_setting add column daily_goal int not null default 10;
alter table user_setting add column timezone varchar(64) not null default 'Asia/Tokyo';
