create table user_tg
(
  id serial not null
    constraint user_tg_pkey
    primary key,
  external_id integer not null,
  user_photo  varchar(255),
  user_photo_id varchar(100),
  created_at  timestamp with time zone,
  updated_at  timestamp with time zone,
  constraint user_tg_key unique(external_id, user_photo, user_photo_id)
);

