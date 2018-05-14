create table bot
(
  id         serial not null
    constraint bot_pkey
    primary key,
  client_id  text,
  token      text,
  name       text,
  created_at timestamp with time zone,
  updated_at timestamp with time zone,
  active     boolean
);

