create table mapping
(
  id        serial not null
    constraint mapping_pkey
    primary key,
  site_code text,
  bot_id    text
);

