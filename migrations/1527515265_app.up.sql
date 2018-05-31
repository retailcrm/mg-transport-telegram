alter table connection
  add constraint connection_key unique (client_id, mg_token),
  alter column client_id set not null,
  alter column api_key set not null,
  alter column api_url set not null,
  alter column mg_url set not null,
  alter column mg_token set not null,
  alter column api_url type varchar(255),
  alter column mg_url type varchar(255);

alter table bot
  add column connection_id integer;

update bot b
  set connection_id = c.id
  from connection c
  where b.client_id = c.client_id;

alter table bot
  drop column client_id,
  alter column channel set not null,
  alter column token set not null,
  add constraint bot_key unique (channel, token);

alter table bot add foreign key (connection_id) references connection on delete cascade;

drop table mapping;
