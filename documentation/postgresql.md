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