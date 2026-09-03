# Ćwiczenia Go — podstawy współbieżności, slices, testing, lintery

> Utworzono: 2026-08-12. Rozwiązujemy krok po kroku, zadanie po zadaniu, w kolejności —
> każde następne zakłada, że poprzednie już „siedzi". Pokrywa się z „Dzień 9" z
> [[PLAN-COVERON-SPRINT.md]] (worker pool, rate limiter, LRU/TTL cache), ale zaczynamy niżej —
> od fundamentów, bo o nie też pytają na Mid/Senior Go (np. `nil` slice vs pusty, `append`
> a współdzielony underlying array, closure nad zmienną pętli).

## Zasady pracy

- Każde zadanie rozwiązujemy w osobnym pliku/pakiecie pod `nordsecurity/golang-exercises/`
  (jeden moduł `go.mod`, żeby nie mnożyć plików jak w `my-go-topics/`).
- Do każdego zadania z współbieżnością: **test + `go test -race`** zanim uznamy za zaliczone.
- Gdzie to ma sens: table-driven test (`t.Run`), nie pojedyncze `assert`y w `main()`.
- Na koniec każdego bloku: `go vet ./...` i `golangci-lint run` — czytamy, co linter
  faktycznie wykrył (nie tylko naprawiamy w ciemno).
- Ja proponuję krótkie zadanie → Ty piszesz kod/rozwiązanie → razem robimy review jak
  na rozmowie („dlaczego tu race?", „co się stanie jeśli usuniesz `defer wg.Done()`?").

---

## Blok 0 — Tablice, slice, mapy (rozgrzewka, ale realne pułapki na rozmowie)

> Status: 1 ✅ zaliczone (array kopiowane przez wartość, slice przez header ze
> wskaźnikiem — potwierdzone testem `-race`). 2 w trakcie.

1. `array` vs `slice` — zadeklaruj oba, pokaż różnicę w przekazywaniu do funkcji
   (kopia całości vs header ze wskaźnikiem na underlying array).
2. `nil` slice vs pusty slice (`[]int{}`) — `len`, `cap`, `== nil`, zachowanie w JSON
   (`json.Marshal`: `null` vs `[]`).
3. Pułapka `append` — dwa slice'y wycięte z tego samego underlying array,
   `append` do jednego nadpisuje dane drugiego. Pokazać kiedy się to dzieje i jak
   tego uniknąć (`copy`, trzeci parametr w `slice[a:b:c]` czyli capping cap).
4. `copy()` a płytka kopia — slice struktur/wskaźników vs slice wartości.
5. Usuwanie elementu ze slice bez zachowania kolejności (swap+truncate) i z zachowaniem
   kolejności (`append(s[:i], s[i+1:]...)`) — złożoność, kiedy używać którego.
6. `new` vs `make` — `new([]int)` zwraca `*[]int` wskazujący na `nil` slice (rzadko
   przydatne), `make([]int, 10, 100)` zwraca gotowy do użycia slice z realną
   pojemnością. Test pokazujący że po samym `new` trzeba i tak dociągnąć `make`,
   zanim cokolwiek dołożysz.
7. Zero value is useful — typ z polem `sync.Mutex` i polem `map[string]int`:
   `var c Counter` działa od razu dla muteksu (zero value = unlocked), ale **nie**
   dla mapy (zero value = `nil` — czytanie OK, zapis panikuje). Dobra okazja żeby
   zderzyć „zero value ready to use" (Effective Go, `bytes.Buffer`/`sync.Mutex`)
   z pułapką nil map z zadania 13.
8. Composite literals — przepisz boilerplate w stylu starego `os.NewFile`
   (`f := new(File); f.fd = fd; f.name = name; ...`) na `return &File{fd: fd, name: name}`.
   Pokaż, że `new(T)` i `&T{}` dają ten sam efekt dla zerowego przypadku.
9. Array przekazywany jako `*[3]float64` (wskaźnik) zamiast przez wartość —
   zaimplementuj `Sum(a *[3]float64) float64` i pokaż, że tu (w przeciwieństwie
   do zadania 1) mutacja przez wskaźnik jest widoczna u wywołującego. Kiedy to
   w ogóle ma sens vs po prostu użycie slice (spoiler: prawie nigdy, ale trzeba
   umieć wytłumaczyć różnicę).
10. Własna implementacja `Append(slice, data []byte) []byte` (jak w Effective Go) —
    sprawdzenie czy `len(slice)+len(data) > cap(slice)`, realokacja przez `make`,
    `copy`, zwrócenie nowego slice header. Wymusza zrozumienie, dlaczego wbudowany
    `append` **musi** zwracać nowy slice, a nie mutować w miejscu.
11. Slice 2D — `[][]byte` z wierszami o różnej długości (jak `LinesOfText` w
    Effective Go) kontra jedna alokacja `pixels := make([]uint8, X*Y)` pocięta na
    wiersze pętlą `picture[i], pixels = pixels[:X], pixels[X:]`. Kiedy wybrać które
    podejście (czy wiersze będą rosły/malały niezależnie, czy nie).
12. Mapy — podstawy: literał złożony, odczyt nieistniejącego klucza zwraca zero
    value typu wartości, `comma-ok` idiom (`v, ok := m[k]`) żeby odróżnić „brak
    klucza" od „wartość zerowa".
13. `nil` map — `var m map[string]int` (bez `make`): odczyt działa i zwraca zero
    value, zapis (`m["x"] = 1`) panikuje. Test z `defer recover()`, który łapie
    ten konkretny panic i asercją potwierdza że wystąpił.
14. `delete(m, k)` jest bezpieczne nawet gdy `k` nie istnieje w mapie — krótki
    sanity test.
15. Dlaczego `slice` nie może być kluczem mapy (equality niezdefiniowana dla
    slice'ów), a `array` czy porównywalny `struct` — może. Spróbuj (i zobacz błąd
    kompilacji) `map[[]int]string{}`, potem zrób to samo z `[3]int` jako kluczem.
16. Współbieżny zapis do tej samej mapy z dwóch goroutine bez synchronizacji →
    `fatal error: concurrent map writes` — to nie jest data race wykrywalny
    wyłącznie przez `-race`, Go ma na to wbudowany check w runtime mapy. Zapowiedź
    Bloku 1 (synchronizacja).

## Blok 1 — Goroutines i `sync`

18. Podstawowy `WaitGroup` — N goroutine, każda inkrementuje współdzielony licznik
    bez muteksu → uruchomić z `-race`, zobaczyć raport, potem naprawić `sync.Mutex`.
19. **Classic bug**: closure nad zmienną pętli `for i := range` przekazaną do goroutine
    bez przekazania jako argument — co drukuje, dlaczego (Go ≤1.21 vs 1.22+ semantyka
    pętli się zmieniła — sprawdzimy który Go macie zainstalowany).
20. `sync.RWMutex` — cache do odczytu/zapisu, wielu czytających jednocześnie,
    jeden piszący na wyłączność. Zmierzyć/pokazać różnicę względem zwykłego `Mutex`.
21. `sync.Once` — leniwa inicjalizacja singletona, bezpieczna współbieżnie.
22. `atomic` (`sync/atomic`, `atomic.Int64` z nowszego Go) — licznik bez muteksu,
    porównanie z zadaniem 18.
23. Deadlock na własne życzenie — dwa muteksy blokowane w odwrotnej kolejności
    przez dwie goroutine, zobaczyć jak Go to raportuje (`fatal error: all goroutines
    are asleep - deadlock!` albo faktyczne zawieszenie).

## Blok 2 — Channels

24. Unbuffered vs buffered channel — blokowanie przy wysyłaniu/odbiorze, kiedy który
    wybrać.
25. `close()` + `range` po kanale — kto zamyka kanał (producent, nigdy konsument),
    co się dzieje przy wysłaniu do zamkniętego kanału (`panic`) i odbiorze
    (`v, ok := <-ch`).
26. `select` z `default` (non-blocking) i z `time.After` (timeout na pojedynczej
    operacji na kanale).
27. Fan-out / fan-in — N workerów czytających z jednego kanału wejściowego,
    wyniki zbierane do jednego kanału wyjściowego, `WaitGroup` + zamknięcie
    kanału wyjściowego po zakończeniu wszystkich workerów.
28. Pipeline — 2-3 etapy połączone kanałami (generator → transformacja → sink),
    z propagacją anulowania przez `context.Context` (nie przez osobny kanał `done`).

## Blok 3 — Context i wzorce anulowania

29. `context.WithCancel` / `WithTimeout` / `WithDeadline` — goroutine, która
    powinna przerwać pracę na `ctx.Done()`, test sprawdzający że faktycznie
    kończy w rozsądnym czasie (nie „na oko" przez `time.Sleep`, tylko przez
    kanał sygnalizujący zakończenie).
30. Propagacja `context.Context` przez kilka warstw funkcji + `context.WithValue`
    (i dlaczego w Nordzie/produkcyjnym kodzie unika się nadużywania `WithValue`).

## Blok 4 — Wzorce z realnych rozmów (łączy wszystko powyżej)

> **Powtórka (jak wracać do tego bloku):** nie podglądać starego rozwiązania
> przed próbą. Pisać od zera, na czas (~20-25 min/wzorzec), mówiąc na głos co
> się robi — dopiero potem porównać z tym co już jest zaimplementowane.
> Oceniać się po tym czy działa i czy umiesz uzasadnić decyzje (nie po tym
> czy kod jest identyczny). Blisko rozmowy: drugie, szybsze przejście "na
> sucho" (opowiedzenie wzorca bez pisania kodu) jako finalny refresh.

31. **Worker pool z ograniczoną liczbą workerów + timeout per zadanie + graceful
    shutdown** przez `signal.NotifyContext` — dokładnie to, co jest w
    [[PLAN-COVERON-SPRINT.md]] jako „Dzień 9".
32. **Rate limiter** (token bucket, `time.Ticker` albo `golang.org/x/time/rate`
    od zera bez biblioteki) — limit N żądań/sekundę, test sprawdzający że
    (N+1)-sze żądanie w tej samej sekundzie czeka/odrzuca.
33. **LRU cache z TTL** — `map` + lista dwukierunkowa (albo `container/list`),
    `sync.Mutex` do bezpieczeństwa współbieżnego, eviction po TTL i po
    przekroczeniu pojemności.
34. Retry z exponential backoff + jitter — funkcja generyczna `Retry(ctx, fn, opts)`,
    test że po N nieudanych próbach zwraca błąd, a nie wisi w nieskończoność.

## Blok 5 — Testowanie (w tym współbieżne)

35. Table-driven tests — `t.Run` z podtestami, `t.Parallel()` gdzie bezpiecznie.
36. Testowanie kodu współbieżnego bez `time.Sleep`: synchronizacja przez kanały,
    `sync.WaitGroup` w teście, albo `testing/synctest` (Go 1.24+, jeśli macie
    tę wersję) do deterministycznego testowania czasu.
37. `go test -race ./...` jako obowiązkowy krok — celowo wprowadzić data race
    w jednym z wcześniejszych zadań i zobaczyć, że test go łapie.
38. Testy z timeoutem (`go test -timeout 5s`) — jak wygląda output przy
    zawieszonej goroutine (deadlock/leak) i jak to zdiagnozować.
39. (Opcjonalnie) Wykrywanie **goroutine leaks** — ręcznie (porównanie
    `runtime.NumGoroutine()` przed/po) albo `go.uber.org/goleak`.

## Blok 6 — Lintery i statyczna analiza

40. `go vet ./...` — co wykrywa (np. błędne `Printf` verbs, `sync.Mutex`
    kopiowany przez wartość).
41. `golangci-lint run` na jednym z modułów w `my-go-topics/` (macie już
    `make lint` w repo) — przejrzeć realne findings, nie tylko `--fix`.
42. Linter jako nauczyciel: specyficznie `govet`'s `copylocks`, `errcheck`
    (niesprawdzone błędy), `staticcheck` (SA-warnings) — po jednym przykładzie
    złego kodu na regułę, żeby umieć to wytłumaczyć słowami na rozmowie.

---

## Kolejność na dziś

Proponuję zacząć od **Bloku 0 (zadania 1-3)**, bo to najszybciej pokazuje, czy
intuicja co do slice'ów jest solidna, a potem przejść do **Bloku 1** (goroutines +
race detector) — to jest serce większości pytań o współbieżność na rozmowach Go.
Mów, jeśli wolisz inną kolejność albo chcesz przeskoczyć od razu do Bloku 4
(wzorce z rozmów).

Powiązane: [[PLAN-COVERON-SPRINT.md]], [[NORDSECURITY-CHAT.md]], [[PLAN_NAUKI.md]].
claude --resume d91f447b-0856-42f8-9b16-f9850e93ae14