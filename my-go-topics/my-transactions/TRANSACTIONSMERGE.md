# MERGE
## Chat
```
MERGE INTO nazwa_tabeli_docelowej AS t
USING (zapytanie_lub_nazwa_tabeli) AS s
ON warunek_porównania
WHEN MATCHED [ AND warunek_dodatkowy ] THEN
    UPDATE SET …
WHEN MATCHED [ AND warunek_dodatkowy ] THEN
    DELETE
WHEN NOT MATCHED [ AND warunek_dodatkowy ] THEN
    INSERT (kol1, kol2, …) VALUES (…)
```

Jak to działa krok po kroku?

Łączenie źródła z tabelą docelową
PostgreSQL wykonuje JOIN między danymi z USING a tabelą docelową na podstawie warunku z ON. Dla każdego wiersza źródła powstaje tzw. candidate change row. 
PostgreSQL

Określenie statusu każdego wiersza:

MATCHED — jeżeli istnieje odpowiadający wiersz w tabeli docelowej

NOT MATCHED BY TARGET — źródło, które nie ma dopasowania w tabeli

NOT MATCHED BY SOURCE — wiersze tabeli, których nie ma w źródle (rozszerzenie SQL) 
PostgreSQL

Ocena kolejnych klauzul WHEN
Dla każdego wiersza PostgreSQL sprawdza w kolejności wszystkie WHEN … THEN i wykonuje tylko pierwszą pasującą akcję

### Przykład
Załóżmy tabelę users i dane z pliku lub innej tabeli updates:
```
CREATE TABLE users (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT        NOT NULL,
    email       TEXT        NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```
```
ALTER TABLE users ALTER COLUMN id DROP IDENTITY;
```

```
INSERT INTO users(name,email,created_at,updated_at) VALUES ('a', 'a@gmail.com', now(), now());
INSERT INTO users(name,email,created_at,updated_at) VALUES ('b', 'b@gmail.com', now(), now());
```

```
CREATE TABLE updates (
    id     BIGINT,
    name   TEXT,
    email  TEXT
);
```
```
INSERT INTO updates(id,name,email) VALUES (1,'User x', 'user.x@gmail.com');
INSERT INTO updates(id,name,email) VALUES (3,'User z', 'user.z@gmail.com');
```

```
MERGE INTO users AS u
USING updates AS s
ON u.id = s.id
WHEN MATCHED THEN
  UPDATE SET name = s.name, email = s.email
WHEN NOT MATCHED THEN
  INSERT (id, name, email) VALUES (s.id, s.name, s.email);

```

Co się dzieje?
Jeśli id z updates istnieje w users → aktualizuje dane.
Jeśli id nie istnieje w users → wstawia nowy wiersz.

## Moje rozkminy
### wiersz target sie zmienil (update)
MERGE INTO users AS u
USING updates AS s
ON u.id = s.id
WHEN MATCHED AND u.name = 'User x' THEN
    UPDATE SET name = s.name
WHEN MATCHED THEN
    UPDATE SET name = s.name, email = s.email
WHEN NOT MATCHED THEN
  INSERT (id, name, email) VALUES (s.id, s.name, s.email);

Jezeli u.id = s.id
i rownolegla transakcja zaktualizowala u.id
wtedy, jak rozumiem, MERGE sie blokuje na tym wierszu?
Potem sprawdza wszystkie argumenty jeszcze raz?

Czyli jezeli rownolegla transakcja usunela wiersz, bedzie - WHEN NOT MATCHED i INSERT

Czyli jezeli rownolegla transakcja zaktualizowala wiersz, bedzie sprawdzenie warunku i UPDATE 1 lub 2 

Jezeli warunek ON nie byl spelniony, a rownolegla transakcja dodala wiersz z tym id,
aktualna transakcja o tym nie wie i dodaje INSERT.
Jezeli rownolegla transakcja zrobi commit wczesniej, INSERT zwroci blad.
