# Start Transakcje Tutorial
```
make start-k3d
make psql-add-to-cluster
make psql-port-forward
make psql-cli
```

https://chatgpt.com/c/692e1e4c-8008-8326-8f21-afb832b7b8cd

```
CREATE DATABASE simplego;

CREATE TABLE public.products (
  id serial PRIMARY KEY,
  price int
);

INSERT INTO products (price) VALUES (100);


```

# Transakcje: 2. Poziomy izolacji w PostgreSQL

https://www.postgresql.org/docs/current/transaction-iso.html
PostgreSQL wspiera 3 poziomy (z 4 standardowych):

Poziom izolacji, 	Opis,	Co chroni?
## READ COMMITTED (domyślny)	
    Każde zapytanie widzi dane zatwierdzone na chwilę jego wykonania.	
    chroni przed „brudnymi odczytami”, ale nie chroni przed non-repeatable reads i phantom reads.


### Dokumentacja - wyjaśnienie chatgpt - snapshot dla SELECT vs dwa SELECTY
#### Dokumentacja
```
Read Committed is the default isolation level in PostgreSQL. 
When a transaction uses this isolation level, a SELECT query (without a FOR UPDATE/SHARE clause) 
sees only data committed before the query began; it never sees either uncommitted data 
or changes committed by concurrent transactions during the query's execution. 
In effect, a SELECT query sees a snapshot of the database as of the instant the query begins to run. 
However, SELECT does see the effects of previous updates executed within its own transaction, 
even though they are not yet committed. 
Also note that two successive SELECT commands can see different data, 
even though they are within a single transaction, 
if other transactions commit changes after the first SELECT starts and before the second SELECT starts.
```

#### Pytanie o rónicę pomidzy snapshot SELECT vs dwa SELECT
Możesz wyjaśnić mi to:
"it never sesss ... changes committed by concurrent transactions during the query's execution"? 
Jak to się ma do zdania 
"note that two successive SELECT commands can see different data,
even though they are within a single transaction, 
if other transactions commit changes after the first SELECT starts and before the second SELECT starts"?


#### Wyjaśnienie - chatgpt - rónica pomidzy snapshot SELECT vs dwa SELECT

PostgreSQL robi snapshot stanu bazy:
"co było zatwierdzone dokładnie w chwili startu zapytania"
I to jest snapshot tylko dla jednego zapytania, a nie dla całej transakcji.

"it never sees ... changes committed by concurrent transactions during the query's execution"
Tzn.:
Jeśli zapytanie SELECT trwa np. 100 ms
i w czasie jego działania inna transakcja commitnie zmianę…
to SELECT NIE zobaczy tej zmiany, bo snapshot był utworzony wcześniej.
SELECT zawsze widzi konsekwentny stan bazy na początku swojego działania i tego się trzyma.


#### Wyjaśnienie - SELECT sees updates not commited within its own transaction
'However, SELECT does see the effects of previous updates executed within its own transaction, 
even though they are not yet committed.'

Czyli transakcja wewnatrz siebie widzi wszystkie zmiany wprowadzone przez siebie, nawet jezeli nie są zacommitowane.
T1:
```
BEGIN;
select * from products;
 id | price 
----+-------
  2 |   800
  1 |   300
(2 rows)

update public.products set price = 400 where id = 2;

select * from products;
 id | price 
----+-------
  1 |   300
  2 |   400
(2 rows)
```

### Dokumentacja - INSERT ON CONFLICT DO UPDATE dla READ COMMITED

"INSERT with an ON CONFLICT DO UPDATE clause behaves similarly. 
In Read Committed mode, each row proposed for insertion will either insert or update. 
Unless there are unrelated errors, one of those two outcomes is guaranteed. 
If a conflict originates in another transaction whose effects are not yet visible to the INSERT, 
the UPDATE clause will affect that row, 
even though possibly no version of that row is conventionally visible to the command."

Czyli dla tabeli products:

T1:
```
BEGIN
INSERT INTO products (id, price) VALUES (3, 1000); 
```

T2:
```
BEGIN;

INSERT INTO products (id, price)
VALUES (3, 500)
ON CONFLICT (id)
DO UPDATE SET price = EXCLUDED.price;
```

T1
```
COMMIT;
```

T2
```
COMMIT; // zrobi update price na 500
```

#### EXCLUDED
Co oznacza EXCLUDED.price?
EXCLUDED to specjalna tabela logiczna zawierająca dane z VALUES(...).
Czyli:
EXCLUDED.id → wartość próbowana do wstawienia
EXCLUDED.price → cena z nowego insertu

#### EXCLUDED - najczęściej uzywana wersja:
Najczęściej używana wersja:
Update tylko wtedy, gdy wartość się różni:
```
INSERT INTO products (id, price)
VALUES (3, 500)
ON CONFLICT (id)
DO UPDATE SET price = EXCLUDED.price
WHERE products.price IS DISTINCT FROM EXCLUDED.price;
```

### Dokumentacja - MERGE
```
MERGE allows the user to specify various combinations of INSERT, UPDATE and DELETE subcommands. 

A MERGE command with both INSERT and UPDATE subcommands looks similar to INSERT with an ON CONFLICT DO UPDATE 
clause but does not guarantee that either INSERT or UPDATE will occur. 

If MERGE attempts an UPDATE or DELETE and the row is concurrently updated but the join condition still passes 
for the current target and the current source tuple, 
then MERGE will behave the same as the UPDATE or DELETE commands 
and perform its action on the updated version of the row. 
However, because MERGE can specify several actions and they can be conditional, 
the conditions for each action are re-evaluated on the updated version of the row, 
starting from the first action, even if the action that had originally matched appears later in the list of actions. 

On the other hand, if the row is concurrently updated so that the join condition fails, 
then MERGE will evaluate the command's NOT MATCHED BY SOURCE and NOT MATCHED [BY TARGET] actions next, 
and execute the first one of each kind that succeeds. 

If the row is concurrently deleted, 
then MERGE will evaluate the command's NOT MATCHED [BY TARGET] actions, 
and execute the first one that succeeds. 

If MERGE attempts an INSERT and a unique index is present and a duplicate row is concurrently inserted, 
then a uniqueness violation error is raised; 
MERGE does not attempt to avoid such errors by restarting evaluation of MATCHED conditions.
```

Rozumiem, ze dla MERGE sprawdzane sa warunki wystepujace w ON.
Jezeli warunek jest spelniony, a wiersz jest modyfikowany przez inna transakcje, jest re-ewaluacja.
Jezeli warunek nie jest spelniony, a rownolegla transakcja zmowyfikuje tabele (np. zrobi insert),
to nie ma re-ewaluacji?
Czyli moze byc tak,

#### TODO - ogarnac komende MERGE (NOT MATCHED BY SOURCE, NOT MATCHED BY TARGET)
#### Chatgpt wytlumaczenie
MERGE:

może mieć wiele akcji (UPDATE, DELETE, INSERT),
każda warunkowa,
i żadna nie jest gwarantowana — musi być ponownie oceniona jeśli wiersz zmienił się równolegle.

##### MERGE — zasada pod konkurencją

Kiedy MERGE działa na wierszu:
sprawdza MATCHED / NOT MATCHED (czy target istnieje)
wybiera akcję (UPDATE, DELETE, INSERT)
próbuje ją zrobić
W tym czasie wiersz może być:
równolegle UPDATE-owany
równolegle DELETE-owany
równolegle INSERT-owany (konflikt unique)
Wtedy:

###### PRZYPADEK 1 — Wiersz równolegle UPDATE-owany, ale JOIN dalej pasuje

MERGE powtórnie ocenia CAŁĄ listę akcji od początku.

Dlaczego?
Bo warunki mogły się zmienić.

##### PRZYPADEK 2 — Wiersz równolegle UPDATE-owany i JOIN JUŻ NIE PASUJE

→ MERGE przechodzi do NOT MATCHED BY SOURCE
lub NOT MATCHED akcji i próbuje je wykonać.

##### PRZYPADEK 3 — Wiersz równolegle DELETE-owany

→ MERGE traktuje jak NOT MATCHED BY TARGET
i wykonuje pierwszą pasującą akcję z tej sekcji.

##### PRZYPADEK 4 — INSERT powoduje conflict unikalny

→ MERGE rzuca błąd,
nie robi automatycznego retry,
nie przechodzi do UPDATE.

To jest największa różnica w stosunku do INSERT ON CONFLICT.

##### Przykład
```
CREATE TABLE products (
  id   INT PRIMARY KEY,
  name TEXT,
  price INT
);
```
```
INSERT INTO products (id,name,price) VALUES (1, 'Phone', 700)
```
Merge:
```
MERGE INTO products p
USING (VALUES (1, 'Phone', 800)) AS s(id, name, price)
ON p.id = s.id
WHEN MATCHED AND s.price > p.price THEN
    UPDATE SET price = s.price           -- akcja 1
WHEN MATCHED THEN
    DELETE                                -- akcja 2
WHEN NOT MATCHED THEN
    INSERT (id, name, price)
    VALUES (s.id, s.name, s.price);       -- akcja 3
```

###### PRZYPADEK 1 — MERGE chce UPDATE, ktoś zmienił wiersz, ale JOIN dalej pasuje

T1 (MERGE) zaczyna:

warunek: s.price > p.price
→ 800 > 700 → pasuje → UPDATE

ale zanim zrobi UPDATE:

T2:
```
UPDATE products SET price = 750 WHERE id=1;
```

Co zrobi MERGE?
blokuje się, czeka
widzi zmieniony wiersz (price 750)
ponownie ocenia warunki od początku:
800 > 750 → nadal pasuje
robi UPDATE do 800

NERW: warunki zawsze oceniane ponownie po kolizji.

###### PRZYPADEK 2 — w trakcie T1 warunek przestaje pasować
T1 (MERGE) zaczyna:

→ chce UPDATE, bo 800 > 700

T2 w międzyczasie:
```
UPDATE products SET price = 900 WHERE id=1;
```

MERGE po odblokowaniu:

widzi nowy price = 900
re-eval całego MERGE:
800 > 900 → ❌ już nie pasuje

przechodzi do kolejnej akcji:
WHEN MATCHED THEN DELETE
DELETE wykona się!
Czyli MERGE może nagle zrobić DELETE zamiast UPDATE.

To jest to, o czym mówi dokumentacja.

###### PRZYPADEK 3 — target został usunięty

T2:
```
DELETE FROM products WHERE id=1;
```

T1 po re-eval:

MATCHED → już nie

przechodzi do:
WHEN NOT MATCHED THEN INSERT

result: MERGE zrobi INSERT

###### PRZYPADEK 4 — równoległy INSERT powoduje conflict
Załóżmy:

T1 (MERGE):
```
INSERT (id,name,price) VALUES (2,'Phone',800)
```

T2 w tej samej chwili:
```
INSERT INTO products VALUES (2,'Phone',999);
```


Efekt:

MERGE rzuca błąd:

ERROR: duplicate key value violates unique constraint "products_pkey"


MERGE nie próbuje przejść do UPDATE,
w przeciwieństwie do:

INSERT ... ON CONFLICT DO UPDATE


## REPEATABLE READ	
    Cała transakcja widzi stały snapshot danych z początku transakcji.	
    chroni przed brudnymi odczytami i non-repeatable reads; 
    w PostgreSQL chroni także skutecznie przed phantom reads dzięki MVCC.

### Dokumentacja
```
UPDATE, DELETE, MERGE, SELECT FOR UPDATE, and SELECT FOR SHARE commands behave the same as SELECT in terms 
of searching for target rows: they will only find target rows that were committed as of the transaction start time. 
However, such a target row might have already been updated (or deleted or locked) by another concurrent transaction by
the time it is found. In this case, the repeatable read transaction will wait for the first updating 
transaction to commit or roll back (if it is still in progress). 
If the first updater rolls back, then its effects are negated and the repeatable read transaction can proceed with 
updating the originally found row. But if the first updater commits (and actually updated or deleted the row, 
not just locked it) then the repeatable read transaction will be rolled back with the message
```

#### Pytanie - w którym momencie transakcja robi snapshot?
Zaraz po 'BEGIN', czy w momencie wykonania pierwszego zapytania?

##### Chat - odpowiedz
W PostgreSQL dla poziomu izolacji REPEATABLE READ 
snapshot jest robiony w momencie wykonania pierwszego zapytania SQL, a nie w chwili BEGIN.

Ale po tym pierwszym zapytaniu snapshot jest już stały aż do COMMIT / ROLLBACK.

BEGIN:
nie tworzy snapshotu
tylko otwiera kontekst transakcji

###### Przykład krok po kroku
🟦 Sesja A (REPEATABLE READ)
BEGIN ISOLATION LEVEL REPEATABLE READ;
-- brak snapshotu


⏳ czekasz 10 sekund...

🟩 Sesja B (w międzyczasie)
INSERT INTO products VALUES (1, 500);
COMMIT;

🟦 Sesja A – pierwsze zapytanie
SELECT * FROM products;


👉 TU powstaje snapshot
👉 nowy wiersz z Sesji B JEST widoczny

🟩 Sesja B ponownie
INSERT INTO products VALUES (2, 600);
COMMIT;

🟦 Sesja A – kolejne zapytanie
SELECT * FROM products;


👉 TEGO wiersza już NIE zobaczy
(snapshot jest zamrożony)

#### Pytanie - co, jezeli update rownoleglej transakcji dotyczy kolumny, ktora nie jest zmieniana w aktualnej?
##### Chat - odpowiedz
Krótka odpowiedź

TAK — w REPEATABLE READ również dostaniesz błąd,
nawet jeśli równoległa transakcja zmienia inną kolumnę tego samego wiersza.

Dlaczego tak się dzieje?
Bo PostgreSQL wersjonuje całe wiersze, a nie kolumny.
UPDATE zawsze tworzy nową wersję CAŁEGO wiersza
nawet jeśli zmieniasz jedną kolumnę
stara wersja wiersza zostaje „zamrożona” w MVCC
W REPEATABLE READ:
Twoja transakcja pracuje na jednym, stałym snapshotcie
nie może „dogadać się” z nową wersją wiersza, której nie widziała

###### Przyklad
🟦 Transakcja T1 (REPEATABLE READ)
BEGIN ISOLATION LEVEL REPEATABLE READ;
SELECT * FROM products WHERE id = 1;
-- widzi: price=100, stock=10

(snapshot zamrożony)

🟩 Transakcja T2 (READ COMMITTED)
BEGIN;
UPDATE products SET stock = 9 WHERE id = 1;
COMMIT;


👉 zmienia inną kolumnę

🟦 T1 próbuje:
UPDATE products SET price = 120 WHERE id = 1;

❌ Wynik:
ERROR: could not serialize access due to concurrent update

## SERIALIZABLE	
    Najsilniejszy — PostgreSQL gwarantuje, że wynik działania jest taki, 
    jakby transakcje były wykonywane jedna po drugiej.


### Dokumentacja
The Serializable isolation level provides the strictest transaction isolation. 
This level emulates serial transaction execution for all committed transactions; 
as if transactions had been executed one after another, serially, rather than concurrently. 
However, like the Repeatable Read level, applications using this level must be prepared to retry transactions 
due to serialization failures.
In fact, this isolation level works exactly the same as Repeatable Read except that it also monitors 
for conditions which could make execution of a concurrent set of serializable transactions behave in a manner 
inconsistent with all possible serial (one at a time) executions of those transactions. 
This monitoring does not introduce any blocking beyond that present in repeatable read, 
but there is some overhead to the monitoring, 
and detection of the conditions which could cause a serialization anomaly will trigger a serialization failure.



# Transakcje: 3. MVCC – jak PostgreSQL to ogarnia

PostgreSQL używa MVCC (Multi-Version Concurrency Control).
Każda zmiana tworzy nową wersję wiersza.

Dlatego SELECT może czytać wersję historyczną, 
a UPDATE może pracować nad aktualną — bez blokowania się nawzajem.

# Transakcje: 4. SELECT FOR UPDATE – do czego służy?
Nikt nie może pisać w te wiersze oprócz mnie.

SELECT ... FOR UPDATE blokuje wybrane wiersze na czas transakcji.

```
BEGIN;
SELECT * FROM products WHERE id = 1 FOR UPDATE;
UPDATE products SET price = prica - 10 WHERE id = 1;
COMMIT;
```
Skutki:
    inne transakcje próbujące zrobić UPDATE/DELETE tego wiersza będą czekać,
    SELECT FOR UPDATE sam w sobie nie blokuje zwykłego SELECT (który czyta snapshot!).

To jest pesymistyczne blokowanie:
    zakładamy, że konflikt na pewno wystąpi,
    więc blokujemy wiersz od razu.

# Transakcje: SELECT FOR SHARE - do czego słuzy
Nikt nie może zmienić danych, które czytam.
Ale inni mogą je również czytać.

# Transakcje: Update 2 wierszy o tym samym id przez rózne transakcje - READ COMMITED
```
CREATE TABLE accounts ( id int PRIMARY KEY, balance int);
INSERT INTO accounts VALUES (1, 1000);
```

I dwie sesje (A i B) próbują wykonać:

```
UPDATE accounts SET balance = balance - X WHERE id = 1;
```

Sesja A:
```
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
```

Co robi PostgreSQL?
Pobiera bieżącą wersję wiersza (balance = 1000).
Tworzy nową wersję wiersza (balance = 900) — MVCC.
Oznacza starą wersję jako „nieważną od tego momentu”.
Zakłada blokadę wiersza typu RowExclusiveLock (dokładnie: FOR UPDATE lock).
Inne transakcje nie mogą aktualizować tego wiersza, dopóki A nie zrobi COMMIT/ROLLBACK.

Sesja B (wykonuje się równocześnie):
```
BEGIN;
UPDATE accounts SET balance = balance - 50 WHERE id = 1;
```
Co się dzieje?

PostgreSQL widzi, że wiersz jest zablokowany przez A.
Sesja B czeka na zwolnienie blokady.
To czekanie może trwać do:
A zrobi COMMIT → B kontynuuje,
A zrobi ROLLBACK → B kontynuuje,
lub czekanie przekroczy lock_timeout.

2. Co dzieje się, gdy Sesja A zrobi COMMIT?

Po COMMIT; w Sesji A:
1. Blokada wiersza zostaje zwolniona.
2. Sesja B nie użyje starej wartości wiersza (1000).  Nigdy nie zaktualizuje "starego snapshotu".
3. Zamiast tego Sesja B:
    - pobiera nową wersję wiersza stworzoną przez A (balance = 900),
    - tworzy kolejną wersję (balance = 850),
    - zakłada blokadę,
wykonuje UPDATE poprawnie.

# Transakcje: SELECT FOR UPDATE

## Transakcje: SELECT FOR UPDATE vs Optimistic Lock
Używaj SELECT FOR UPDATE, gdy:
    zmieniasz licznik/stan magazynowy,
    robisz przelew,
    wykonujesz algorytm, który musi widzieć aktualne dane.

Używaj optimistic locking, gdy:
    większość operacji to CRUD,
    konflikty zdarzają się rzadko,
    chcesz wysokiej skalowalności.

## Transakcje: Moje testy
### Select * for update; vs insert...;
```
# Transakcja A
begin;
select * from products for update;

# Transakcja B
begin;
insert into products (price) values (800);
```
Wszystko ok.
Wsześniej był error, ze względu na błąd w zapytaniu sql.

# Transakcje: Problemy
Zakładam istnienie tabeli products:
```
CREATE DATABASE simplego;

CREATE TABLE public.products (
  id serial PRIMARY KEY,
  price int
);

INSERT INTO products (price) VALUES (100);
```
## Transakcje: Problemy - Lost update – przykład dla READ COMMITTED
### Tłumaczenie: na czym polega "lost update"?

Dwie transakcje czytają ten sam wiersz, obie go modyfikują i zapisują.
Ostatnia wygrywa → pierwsza modyfikacja „znika”.

Przykład:
Transakcja A:
```
BEGIN;
SELECT price FROM products WHERE id=1; -- price = 100
UPDATE products SET price = 120 WHERE id=1;
```

Transakcja B:
```
BEGIN;
SELECT price FROM products WHERE id=1; -- price = 100
UPDATE products SET price = 90 WHERE id=1;
```

price = 90    -- zmiana A została nadpisana

### Dlaczego w PostgreSQL lost-update występuje tylko na READ COMMITTED?

Bo:
PG pozwala dwóm transakcjom czytać tę samą wersję wiersza,
podczas update dopiero wtedy zakłada blokadę,
drugi UPDATE nie widzi, że ktoś wcześniej również oparł się o starą wersję.
To klasyczny lost update.

### Co dzieje się na REPEATABLE READ?
Drugi update powoduje błąd:
```
ERROR: could not serialize access due to concurrent update
```
PostgreSQL na tym poziomie:
śledzi wersje rekordów,
wykrywa, że transakcja bazuje na starej wersji,
musi przerwać transakcję i wymusić retry.
Czyli lost update nie przejdzie.

### Co z SELECT FOR UPDATE?
SELECT ... FOR UPDATE zapobiega "lost update" nawet na READ COMMITTED.

Bo:
transakcja, która pierwsza zrobi SELECT FOR UPDATE, blokuje wiersz,
druga transakcja musi czekać,
nie może przypadkiem nadpisać bazując na starej wersji wiersza.

To jest manualna forma pessimistic locking.

### Co z optimistic locking?
Optymistyczna blokada (najczęściej version lub updated_at) też zapobiega lost-update, bo:
UPDATE zawiera warunek WHERE id=1 AND version=5,
jeśli inna transakcja zmieniła wersję → UPDATE niczego nie modyfikuje,
aplikacja widzi, że trzeba zrobić retry.
To jest mechanizm znany z ORMs (Hibernate, GORM itp.).

### Podsumowanie w jednym zdaniu
Lost update w PostgreSQL jest możliwy na READ COMMITTED, ale wyższe poziomy izolacji, SELECT FOR UPDATE i optimistic locking skutecznie mu zapobiegają.

### Róznica dla set price = x vs set price = price - x w 'READ COMMITED' (moje testy)
Z testow wynika, ze jezeli obie transakcje maja
```
set price = price - x
```
to wynik jest poprawny (uwzglednia obie transakcje)
Jezeli obie maja:
```
set price = x
```
to wynik zostaje nadpisany wartoscia ostatniej transakcji

### Róznica dla set price = x vs set price = price - x w 'REPETABLE READ' (moje testy)
Nie testowane - ale wydaje mi sie, ze powinien byc blad - bo update bazuje na niaktualnej wersji rekordu.

## Transakcje: Problemy - Non-repetable read – przykład dla READ COMMITTED
Definicja:
Transakcja A czyta ten sam wiersz dwa razy, ale drugi odczyt widzi inną wartość, bo transakcja B w międzyczasie zrobiła commit.


Transakcja A:
```
BEGIN;
-- 1. Odczyt
SELECT price FROM products WHERE id = 1;  
-- wynik: 100
```

Transakcja B:
```
BEGIN;
UPDATE products SET price = 200 WHERE id = 1;
COMMIT;
```

Transakcja A:
```
-- 2. Drugi odczyt tego samego wiersza
SELECT price FROM products WHERE id = 1;
-- wynik: 200  <-- INNE niż wcześniej!
```

### Dlaczego tak się dzieje?

Bo w READ COMMITTED każde osobne zapytanie widzi najnowsze zatwierdzone dane.
Transakcja A nie ma stabilnego snapshotu — tylko pojedyncze SELECT-y są spójne.


## Transakcje: Problemy - Phantom read - przykład dla READ COMMITTED
### Definicja Phantom read:
Transakcja A wykonuje dwa razy SELECT z warunkiem, który obejmuje wiele wierszy.
Między odczytami transakcja B dodaje/usunie wiersze → wynik SELECT zmienia się.

```
CREATE TABLE orders (
  id serial PRIMARY KEY,
  amount int
);

INSERT INTO orders (amount) VALUES (10), (20);
```

Transakcja A:
```
BEGIN;
SELECT COUNT(*) FROM orders WHERE amount > 0;
-- wynik: 2
```

Transakcja B:
```
BEGIN;
INSERT INTO orders (amount) VALUES (30);
COMMIT;
```

Transakcja A:
```
SELECT COUNT(*) FROM orders WHERE amount > 0;
-- wynik: 3  <-- DODATKOWY WIERSZ (phantom)
```

### Dlaczego phantom read występuje?

Bo PostgreSQL w READ COMMITTED nie tworzy snapshotu trwającego przez całą transakcję.
Każdy SELECT ma własny snapshot → więc SELECT-y w tej samej transakcji mogą widzieć inne zestawy wierszy.

### Co się stanie na REPEATABLE READ?

Oba problemy znikają:

non-repeatable read: drugi SELECT zwróci tę samą wartość, nawet jeśli inna transakcja zrobi commit → zobaczysz starą wersję.

phantom read: PostgreSQL nie pokazuje nowych wierszy dodanych po rozpoczęciu transakcji.

## Transakcje: Problemy - Serialization anomaly (TODO - jakie poziomy?)
Serialization anomaly (anomalia serializacji) to sytuacja, w której wynik wykonania grupy transakcji równoległych nie jest równoważny żadnemu możliwemu wykonaniu sekwencyjnemu (serialnemu).

Innymi słowy:

Transakcje razem zrobiły coś, czego nie da się odtworzyć, gdyby wykonać je jedna po drugiej.

To jest najpoważniejszy rodzaj problemów w izolacji transakcji — oznacza, że baza danych nie zachowała poprawnej izolacji logicznej.

## Transakcje: Problemy REPETABLE READ - 
### Transakcje: Problemy REPETABLE READ - problem opisany w dokumentacji - delete dla website:
Możesz wyjaśnić na przykładzie tą częśc dokumentacji postgresql 
(https://www.postgresql.org/docs/current/transaction-iso.html):

More complex usage can produce undesirable results in Read Committed mode. 
For example, consider a DELETE command operating on data that is being both 
added and removed from its restriction criteria by another command, 
e.g., assume website is a two-row table with website.hits equaling 9 and 10:

BEGIN;
UPDATE website SET hits = hits + 1;
-- run from another session:  DELETE FROM website WHERE hits = 10;
COMMIT;
The DELETE will have no effect even though there is a website.hits = 10 
row before and after the UPDATE. 
This occurs because the pre-update row value 9 is skipped, 
and when the UPDATE completes and DELETE obtains a lock, 
the new row value is no longer 10 but 11, which no longer matches the criteria.

### Wytlumaczenie:
#### Sytuacja poczatkowa
Tabela 'website'
| id | hits |
| -- | ---- |
| 1  | 9    |
| 2  | 10   |


#### S1:
```
BEGIN;
UPDATE website SET hits = hits + 1;
```

To oznacza:
| id | old hits | new hits |
| -- | -------- | -------- |
| 1  | 9        | 10       |
| 2  | 10       | 11       |

Ale S1 jeszcze nie zrobiła COMMIT, więc inne transakcje nie widzą tych zmian.

#### S2: (w tym czasie uruchamia)
```
BEGIN
DELETE FROM website WHERE hits = 10;
```
Co S2 widzi?
Ponieważ S1 jeszcze nie commitowała:
| id | hits widoczne w S2 |
| -- | ------------------ |
| 1  | 9                  |
| 2  | 10                 |

DELETE działa tak:

1. Skanuje tabelę na żywo (READ COMMITTED)
2. Szuka wierszy o hits = 10
3.  Trafia na:

| id=1 → hits=9 → NIE pasuje → pomija
| id=2 → hits=10 → PASUJE → próbuje usunąć

ALE…

Aby usunąć wiersz, musi zdobyć lock na wierszu.
Idzie więc po wierszu:
S2 próbuje zlockować wiersz id=2
ALE wiersz jest aktualizowany przez S1 (UPDATE website SET hits = hits + 1)
S2 czeka aż S1 skończy UPDATE

#### Co dzieje się po stronie S1?
Kiedy S1 wykonuje:
```
COMMIT;
```
Zmiany staja sie widoczne:
| id | hits |
| -- | ---- |
| 1  | 10   |
| 2  | 11   |

#### Co robi S2 po odblokowaniu wiersza?
Po tym jak S1 zwolni locki, S2 dostaje lock na wierszu ale musi sprawdzić, czy warunek wciąż pasuje.

To jest bardzo ważne:
PostgreSQL nie usuwa wiersza, jeśli po odblokowaniu nie spełnia już warunku WHERE.

Wiersz id=2 W TEJ CHWILI MA → hits = 11

Już nie pasuje do WHERE hits = 10.

#### ❌ Wniosek: DELETE niczego nie usuwa

Mamy “okno czasowe”, w którym:
W momencie skanowania tabela miała hits = 10
Ale w momencie uzyskania locka wartość była już 11
Więc DELETE nie ma prawa tego ruszyć

#### Jak to naprawić?
##### 1) Użyj REPEATABLE READ

Zapytanie DELETE zobaczy stabilny snapshot.

BEGIN ISOLATION LEVEL REPEATABLE READ;
DELETE FROM website WHERE hits = 10;
COMMIT;


Wtedy:
widzi wartość 10
próbuje skasować
po odblokowaniu wiersza okazuje się, że wartość to 11
→ PostgreSQL rzuci serialization failure
I to jest poprawne zachowanie.

##### 2) SELECT FOR UPDATE przed DELETE
```
BEGIN;
SELECT id FROM website WHERE hits = 10 FOR UPDATE;
DELETE FROM website WHERE hits = 10;
COMMIT;
```
Gdy S1 próbuje zmienić wiersz, zatrzyma się — bo FOR UPDATE trzyma lock.

#### Przykład wystąpienia problemu w tabeli website dla READ COMMITED:
```
CREATE TABLE website (
  id serial PRIMARY KEY,
  hits integer
);

INSERT INTO website (hits) VALUES (9), (10);
```

S1:
```
BEGIN;
UPDATE website SET hits = hits + 1;
```
To aktualizuje:
id=1, 9 → 10
id=2, 10 → 11
ALE trzyma blokady do czasu COMMIT, więc inne transakcje muszą czekać.

S2:
```
DELETE FROM website WHERE hits = 10;
```
S2 widzi:
| id | hits |
| -- | ---- |
| 1  | 9    |
| 2  | 10   |

Więc próbuje usunąć id=2, ale…
musi zdobyć blokadę, która jest trzymana przez S1.


# TODO - Transakcje: SSI (Serializable Snapshot Isolation)
```
CREATE TABLE public.doctors ( id serial PRIMARY KEY, name varchar(255), on_call boolean);

INSERT INTO doctors (name, on_call) values ('dr. House', TRUE);
INSERT INTO doctors (name, on_call) values ('dr. Who', TRUE);
```

T1, T2:
```
BEGIN;
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
```

T1, T2:
```
SELECT * FROM doctors WHERE id IN (1,2);
```

T1:
```
UPDATE doctors SET on_call=FALSE WHERE id=1;
```

T2:
```
UPDATE doctors SET on_call=FALSE WHERE id=2;
```

T1
```
COMMIT;
> COMMIT
```

T2
```
COMMIT;
>ERROR:  could not serialize access due to read/write dependencies among transactions
>DETAIL:  Reason code: Canceled on identification as a pivot, during commit attempt.
>HINT:  The transaction might succeed if retried. 
```

## Rozumiem, ze zaleznosc wykrywa przez SELECT? 

Rozumiem, ze SELECT jest potrzebny, zeby wywołać serializację:

Serializable Snapshot Isolation (SSI) wykrywa anomalie na podstawie:
konfliktów read → write (najważniejsze)
konfliktów write → read
konfliktów write → write (rzadziej)

## TODO Jakie konflikty wykrywa Postgresql
### 4. Jakie konflikty wykrywa PostgreSQL SSI?
1️⃣ Read-write conflicts

A czyta coś, co potem B zmienia.

2️⃣ Write-write conflicts

A i B próbują zmienić ten sam wiersz (to PG blokuje nawet w Read Committed).

3️⃣ Predicate conflicts (phantoms)

A robi:

SELECT * FROM accounts WHERE balance > 0;


PG zapamiętuje ten warunek (predicate lock).

Jeśli B potem wstawi coś nowego, co spełnia ten warunek → konflikt.

### 6. W skrócie: co robi PostgreSQL?

Można to zamknąć w 3 punktach:

Każda transakcja dostaje snapshot (jak w REPEATABLE READ).

Baza śledzi dwu– i trzy-transakcyjne zależności (A czyta po B, B pisze po C itd.).

Jeśli wykryje cykl zależności → ROLLBACK jednej transakcji przy COMMIT.

Dzięki temu wynik działania transakcji zawsze da się ułożyć w jakąś kolejność serialną.
## TODO - write scew
Rozważ sytuację:

Transakcja A czyta dane X i Y

Transakcja B czyta dane X i Y

A aktualizuje X

B aktualizuje Y

W normalnym Repeatable Read obie transakcje commitują → write skew.

W SERIALIZABLE dzieje się:

🔹 Krok 1: PostgreSQL widzi, że A i B przeczytały te same dane

Tworzy zależności „read dependency”.

🔹 Krok 2: PostgreSQL widzi, że A i B zmodyfikowały różne wiersze

Tworzy zależności „write dependency”.

🔹 Krok 3: Powstaje cykl zależności:
A → B → A


Cykliczna zależność = serialization anomaly

🔹 Krok 4: PostgreSQL automatycznie ubija jedną transakcję:
ERROR: could not serialize access due to read/write dependencies among transactions
DETAIL: Reason code: Canceled on conflict out


Czyli dopiero przy COMMIT PG stwierdza, że wynik byłby nieserializowalny — więc go nie dopuszcza.

### TODO - write scew
Czy 

### TODO
🔹 przykład write skew realnie działający w PostgreSQL na SERIALIZABLE (z błędem)
🔹 przykład wykrytego phantom read i rollbacku
🔹 diagram jak PostgreSQL wykrywa cykl (A → B → A)

Co wykrywa PostgreSQL SERIALIZABLE (SSI), niezależnie od constraintów?
1. Write skew

Dwie transakcje czytają te same dane, aktualizują różne, razem tworzą nieserializowalny wynik.

2. Phantom anomalies

Transakcja robi SELECT z warunkiem → inna transakcja dodaje rekord, który ten SELECT powinien widzieć → cykl → rollback.

3. Anomalie 3-transakcyjne (rw-dependency cycles)

A czyta po B, B czyta po C, C pisze po A → rollback.

4. Zależności między snapshotami

PG analizuje, co kto czytał i pisał → wykrywa niespójny graf.

Żaden z tych mechanizmów nie wymaga posiadania constraintów.

# TODO - wyjaśnij MVCC
# TODO - czy poziomy izolacji słóza wylacznie do zapobiegania problemom:
READ COMMITED - zapobiega: dirty read
REPETABLE READ - zapobiega: lost update, non-repetable read i phantom read
SERIALIZABLE - zapobiega: serialization anomaly

Czy sluza do czegos jeszcze?


# TODO - wyjasnic
Repetable Read
Jzeli transakcja ma updateowac wiersz, a inna transakcja zrobi to wczesniej (update lub delete),
powinien zostac rzucony blad.
Co w przypadku, gdy update dotyczy innej kolumny?

# TODO - serial number
Przyklad np. zapytnia, gdzie modyfikowany jest typ danych 'serial', ktory powinien byc widoczny w transakcji 'Repetable Read', ale nie powinny byc widoczne zadne inne zmiany



# TODO
Szkolenie finansowane nawet na zewnatrz
Mentoring?