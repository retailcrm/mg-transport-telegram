create table site_bot
(
  id        serial not null
    constraint site_bot_pkey
    primary key,
  site_code text,
  bot_id    text
);

