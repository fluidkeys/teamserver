CREATE TABLE team_users (
  id SERIAL UNIQUE
, team_id INT REFERENCES teams (id) ON UPDATE CASCADE ON DELETE CASCADE
, fingerprint VARCHAR REFERENCES public_keys (fingerprint) ON UPDATE CASCADE
, is_admin BOOLEAN DEFAULT false
, PRIMARY KEY (team_id,fingerprint)
);