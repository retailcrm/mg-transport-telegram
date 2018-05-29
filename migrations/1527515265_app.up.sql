alter table connection
  add constraint connection_client_id_key unique (api_key),
  add constraint connection_api_key_key unique (api_url),
  add constraint connection_api_url_key unique (mg_url),
  add constraint connection_mg_url_key unique (mg_token);
alter table connection
  alter column client_id set not null,
  alter column api_key set not null,
  alter column api_url set not null;

alter table bot
  alter column client_id type integer using client_id::integer,
  alter column token set not null,
  add constraint bot_token_key unique (token);

alter table bot
  rename column client_id to connection_id;

drop table mapping;