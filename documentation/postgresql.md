## Logowanie do psql
psql -h <host> -p <port> -U <user> -d <database>

## Lista schematow
\dn - lista schematow

```
SELECT schema_name
FROM information_schema.schemata
ORDER BY schema_name;
```

## Listowanie tabel ze schematu x:
\dt x.*


## Books
https://chatgpt.com/c/691ccaea-f6a8-8321-ad31-f86cb72db73e


# Utworzenie bazy danych + user
```
psql -U postgres
```
lub:
```
sudo -u postgres psql
```

```
CREATE USER mysqlcuser WITH PASSWORD 'mypassword';
```
lub:
```
CREATE USER mysqlcuser WITH PASSWORD 'mypassword' CREATEDB;
```
```
CREATE DATABASE mysqlcdb;
```
```
GRANT ALL PRIVILEGES ON DATABASE mysqlcdb TO mysqlcuser;
```
```
ALTER DATABASE mysqlcdb OWNER TO mysqlcuser;
```
```
\c mysqlcdb
```
```
GRANT ALL ON SCHEMA public TO mysqlcuser;
```


## Szybka wersja
```
sudo -u postgres psql
```
```
CREATE USER mymigrationsuser WITH PASSWORD 'mypassword';
CREATE DATABASE mymigrationsdb;
ALTER DATABASE mymigrationsdb OWNER TO mymigrationsuser;
GRANT ALL PRIVILEGES ON DATABASE mymigrationsdb TO mymigrationsuser;
```
```
\c mymigrationsdb
```
```
GRANT ALL ON SCHEMA public TO mymigrationsuser;
```