alter table connection
  alter column client_id type text,
  alter column api_key type text,
  alter column api_url type text,
  alter column mg_url type text,
  alter column mg_token type text;

alter table bot
  alter column client_id type text,
  alter column name type text,
  alter column token type text;
