CREATE TABLE team_join_requests (
  id SERIAL UNIQUE
, team_id INT REFERENCES teams (id) ON UPDATE CASCADE ON DELETE CASCADE
, fingerprint VARCHAR REFERENCES public_keys (fingerprint) ON UPDATE CASCADE
, created_at TIMESTAMP
, PRIMARY KEY (team_id,fingerprint)
);