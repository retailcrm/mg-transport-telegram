create table connection
(
  id         serial not null
    constraint connection_pkey
    primary key,
  client_id  text,
  api_key    text,
  api_url    text,
  created_at timestamp with time zone,
  updated_at timestamp with time zone,
  active     boolean
);

