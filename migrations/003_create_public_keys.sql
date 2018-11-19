CREATE TABLE public_keys (
  id SERIAL UNIQUE
, fingerprint VARCHAR PRIMARY KEY
, armoredPublicKey TEXT
);
