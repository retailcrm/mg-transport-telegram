create table users
(
  id serial not null
    constraint users_pkey
    primary key,
  external_id integer not null,
  user_photo_url  varchar(255),
  user_photo_id varchar(100),
  created_at  timestamp with time zone,
  updated_at  timestamp with time zone,
  constraint users_key unique(external_id, user_photo_url, user_photo_id)
);
