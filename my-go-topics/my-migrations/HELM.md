Dockefile
Klaster z rejestrem działają

Wygląda na to, że muszę sealed-secrets robić dla każdego namespace?
odp.: Tak - secrets są unikalne dla namespace.

Wyjaśnienie, czemu bitnami nie utworzył mi bazy danych z użytkownikiem i hasłem:
https://chatgpt.com/g/g-p-6985a4422d4881919dc148b04c697b83-kubernetes/c/69f64d17-5ac8-83eb-a686-8fe07cf264c0

Najpier muszę się zalogować do bazy danych i utworzyć:

```
CREATE USER appuser WITH PASSWORD 'strongpassword';
CREATE DATABASE appdb;
GRANT ALL PRIVILEGES ON DATABASE appdb TO appuser;
```
```
\c appdb
GRANT ALL ON SCHEMA public TO appuser;
ALTER SCHEMA public OWNER TO appuser;
```