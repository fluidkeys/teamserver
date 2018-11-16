CREATE DATABASE teamserver_development;
CREATE USER teamserver WITH ENCRYPTED PASSWORD ':password';
GRANT ALL PRIVILEGES ON DATABASE teamserver_development TO teamserver;
CREATE DATABASE teamserver;