alter table connection
  alter column client_id type varchar(70),
  alter column api_key type varchar(100),
  alter column api_url type varchar(100),
  alter column mg_url type varchar(100),
  alter column mg_token type varchar(100);

alter table bot
  alter column client_id type varchar(70),
  alter column name type varchar(40),
  alter column token type varchar(100);
