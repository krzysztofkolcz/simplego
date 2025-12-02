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

PostgreSQL wspiera 3 poziomy (z 4 standardowych):

Poziom izolacji, 	Opis,	Co chroni?
READ COMMITTED (domyślny)	
    Każde zapytanie widzi dane zatwierdzone na chwilę jego wykonania.	
    chroni przed „brudnymi odczytami”, ale nie chroni przed non-repeatable reads i phantom reads.

REPEATABLE READ	
    Cała transakcja widzi stały snapshot danych z początku transakcji.	
    chroni przed brudnymi odczytami i non-repeatable reads; 
    w PostgreSQL chroni także skutecznie przed phantom reads dzięki MVCC.

SERIALIZABLE	
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

# Transakcje: Update 2 wierszy o tym samym id przez rózne transakcje
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
select * from proucts for update;

# Transakcja B
begin;
insert into products (price) values (800);
ERROR:  current transaction is aborted, commands ignored until end of transaction block
```
Error, z tego co rozumiem dlatego, ze id jest typu 'serial', a select *... for update zakłada tez lock na index?