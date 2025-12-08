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


### SELECT sees updates not commited within its own transaction
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


## REPEATABLE READ	
    Cała transakcja widzi stały snapshot danych z początku transakcji.	
    chroni przed brudnymi odczytami i non-repeatable reads; 
    w PostgreSQL chroni także skutecznie przed phantom reads dzięki MVCC.

## SERIALIZABLE	
    Najsilniejszy — PostgreSQL gwarantuje, że wynik działania jest taki, 
    jakby transakcje były wykonywane jedna po drugiej.

# Transakcje: 3. MVCC – jak PostgreSQL to ogarnia

PostgreSQL używa MVCC (Multi-Version Concurrency Control).
Każda zmiana tworzy nową wersję wiersza.

Dlatego SELECT może czytać wersję historyczną, 
a UPDATE może pracować nad aktualną — bez blokowania się nawzajem.

# Transakcje: 4. SELECT FOR UPDATE – do czego służy?
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

## Transakcje: Problemy - Serialization anomaly
Serialization anomaly (anomalia serializacji) to sytuacja, w której wynik wykonania grupy transakcji równoległych nie jest równoważny żadnemu możliwemu wykonaniu sekwencyjnemu (serialnemu).

Innymi słowy:

Transakcje razem zrobiły coś, czego nie da się odtworzyć, gdyby wykonać je jedna po drugiej.

To jest najpoważniejszy rodzaj problemów w izolacji transakcji — oznacza, że baza danych nie zachowała poprawnej izolacji logicznej.

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