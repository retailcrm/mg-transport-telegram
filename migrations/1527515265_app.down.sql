alter table connection
  drop constraint connection_key,
  alter column client_id drop not null,
  alter column api_key drop not null,
  alter column api_url drop not null,
  alter column mg_url drop not null,
  alter column mg_token drop not null,
  alter column api_url type varchar(100),
  alter column mg_url type varchar(100);

alter table bot
  add column client_id varchar(70);

update bot b
  set client_id = c.client_id
  from connection c
  where b.connection_id = c.id;

alter table bot
  drop column connection_id,
  alter column channel drop not null,
  alter column token drop not null,
  drop constraint bot_key;

create table mapping
(
  id serial not null
    constraint mapping_pkey
    primary key,
  site_code text,
  bot_id    text
);
