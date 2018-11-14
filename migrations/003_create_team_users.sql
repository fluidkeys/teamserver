CREATE TABLE team_users (
  id SERIAL UNIQUE
, team_id INT REFERENCES teams (id) ON UPDATE CASCADE ON DELETE CASCADE
, fingerprint VARCHAR REFERENCES public_keys (fingerprint) ON UPDATE CASCADE
, PRIMARY KEY (team_id,fingerprint)
, CONSTRAINT team_users_pkey PRIMARY KEY (team_id, fingerprint)
);