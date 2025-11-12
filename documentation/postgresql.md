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
