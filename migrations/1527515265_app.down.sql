alter table connection
  drop constraint connection_client_id_key,
  drop constraint connection_api_key_key,
  drop constraint connection_api_url_key,
  drop constraint connection_mg_url_key;
alter table connection
  alter column client_id drop not null,
  alter column api_key drop not null,
  alter column api_url drop not null;

alter table bot
  alter column connection_id type varchar(70),
  alter column token drop not null,
  drop constraint bot_token_key;

alter table bot
  rename column connection_id to client_id;

create table mapping
(
  id        serial not null
    constraint mapping_pkey
    primary key,
  site_code text,
  bot_id    text
);