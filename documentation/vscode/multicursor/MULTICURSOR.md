a jak zaznaczyć multikursorem te same wyrazy, które chce zmienić?

## 🎯 1️⃣ **Zaznacz kolejne wystąpienie wyrazu**

👉 **Skrót:**
`Ctrl + D`  (Windows / Linux)
`Cmd + D`  (Mac)

### 🔹 Jak działa:

1. Ustaw kursor na wyrazie, np. `m`
2. Naciśnij `Ctrl + D` → zaznaczy następne wystąpienie `m`
3. Naciśnij ponownie `Ctrl + D`, by dodać kolejne
4. Piszesz — wszystkie się zmieniają równocześnie

### 💡 Tip:

Jeśli przesadziłeś (zaznaczyłeś za dużo), użyj `Ctrl + U` (Undo last selection)

---

## 🎯 2️⃣ **Zaznacz wszystkie wystąpienia wyrazu w pliku**

👉 **Skrót:**
`Ctrl + Shift + L`
(Mac: `Cmd + Shift + L`)

### 🔹 Jak działa:

* Zaznacza wszystkie wystąpienia bieżącego zaznaczenia.
* Super, gdy chcesz podmienić wszystkie `m.` na `result.A.` jednym ruchem.

---

## 🎯 3️⃣ **Dodawanie kursora ręcznie**

👉 **Skróty:**

* `Alt + Click` → dodaje kursor w dowolnym miejscu (mogę dodać nawet w tej samej linini)

TODO -alt (option) + shift + strzalka -  nie dziala. Powiela linie.
* `Alt + Shift + ↑ / ↓` → dodaje kursor w kolejnych liniach pionowo (tryb kolumnowy)

### 🔹 Użycie:

Idealne przy modyfikacji bloków:

```go
A typeA
B typeB
C typeC
```

Zrób kolumnowy multi-cursor i dopisz `result.` przed każdą linią.

---

## 🎯 4️⃣ **Tryb zaznaczania kolumnowego (Column Selection)**

👉 **Skróty:**

* Windows/Linux: `Alt + Shift + Drag myszą`
* Mac: `Option + Shift + Drag myszą`

### 🔹 Działa idealnie, gdy:

Masz wiele linii i chcesz edytować początek lub koniec każdej:

```go
m.someMethod()
n.someMethod()
o.someMethod()
```

Zaznaczasz pionowo przed `m` → piszesz `result.` i wszystkie linie dostają prefix.

---

## 🎯 5️⃣ **Wyszukiwanie + Edycja**

👉 `Ctrl + F2` — **Select All Occurrences of Find Match**

### 🔹 Działa tak:

1. `Ctrl + F` (`command + f`) → wyszukaj np. `m.`
2. `Alt + Enter` (`option + Enter`) → doda kursor w każdym dopasowaniu
3. Teraz możesz edytować wszystkie wystąpienia naraz

---

## 🧠 Tip dla programisty Go

Jeśli chcesz, by multi-cursor działał tylko na **całych słowach** (a nie np. „m” w środku „some”), włącz opcję:

**Settings → Editor: Find → Match Whole Word**

---

## ⚡ Przykład praktyczny

Masz:

```go
m.some()
n.some()
o.some()
```

### Chcesz:

```go
result.A.some()
result.B.some()
result.C.some()
```

➡️ Ustaw kursor na `m`
➡️ `Ctrl + D` dwa razy, żeby dodać `n` i `o`
➡️ Wpisz `result.`
➡️ Następnie ręcznie zmień litery A/B/C

---

Chcesz, żebym Ci pokazał **mini ćwiczenie w Go**, gdzie możesz przećwiczyć wszystkie te skróty (multi-cursor, column select, replace, itd.) na jednym fragmencie kodu?


Świetny pomysł 💪 — nauka **multi-cursor i regex-edytowania** w VS Code na praktycznych przykładach (zwłaszcza na kodzie Go) da Ci ogromną przewagę przy refaktorach.
Zrobimy **mini-warsztat**: 5 zestawów ćwiczeń (po jednym dla każdej techniki).

---

## 🧩 **Setup (jeden plik do ćwiczeń)**

Utwórz plik `vscode_multicursor_practice.go` i wklej:

```go
package main

import "fmt"

func main() {
    m, n, o, p, q := newX()
    m.doSomething()
    n.doSomething()
    o.doSomething()
    p.doSomething()
    q.doSomething()

    fmt.Println(m, n, o, p, q)
}

func newX() (typeA, typeB, typeC, typeD, typeE) {
    return x, y, z, k, l
}
```

---

# 🧠 **Ćwiczenie 1: `Ctrl + D` – wybieranie kolejnych wystąpień**

### 🎯 Cel:

Zamienić wszystkie `doSomething()` na `doSomethingElse()`.

### 🪄 Kroki:

1. Ustaw kursor na `doSomething`
2. Naciśnij `Ctrl + D` kilka razy (zaznacza kolejne wystąpienia)
3. Wpisz `doSomethingElse`
4. Wszystkie wystąpienia zmienią się jednocześnie

---

# 🧠 **Ćwiczenie 2: `Ctrl + Shift + L` – zaznacz wszystkie wystąpienia**

### 🎯 Cel:

Podmienić wszystkie zmienne `m, n, o, p, q` na `result.A, result.B...` (pierwszy krok refaktoru).

### 🪄 Kroki:

1. Zaznacz `m`
2. `Ctrl + Shift + L` – kursory pojawią się na każdym `m`
3. Wpisz `result.A`
4. Powtórz dla `n`, `o`, `p`, `q`

💡 Tip: działa nawet jeśli zmienne są w różnych częściach pliku.

---

# 🧠 **Ćwiczenie 3: `Alt + Click` / `Alt + Shift + ↑ / ↓` – kolumnowe pisanie**

### 🎯 Cel:

Przed każdą linią dodać `result.`

Masz:

```go
m.doSomething()
n.doSomething()
o.doSomething()
```

### 🪄 Kroki:

1. Ustaw kursor przed `m`
2. Trzymaj `Alt + Shift`, naciśnij `↓` (lub przeciągnij myszką w dół)
3. Pojawią się kursory przed każdą linią
4. Wpisz `result.`
   → Otrzymasz:

```go
result.m.doSomething()
result.n.doSomething()
result.o.doSomething()
```

---

# 🧠 **Ćwiczenie 4: `Alt + Enter` / `Ctrl + F2` – zaznacz wszystkie dopasowania wyszukiwania**

### 🎯 Cel:

Zmienić `fmt.Println` na `log.Println`

### 🪄 Kroki:

1. `Ctrl + F` → wyszukaj `fmt.Println`
Mac: Command + f (???)
2. `Alt + Enter` → doda kursor w każdym dopasowaniu
3. Wpisz `log.Println`

> Idealne, gdy chcesz hurtowo zmienić wywołania tej samej funkcji w całym pliku.

---

# 🧠 **Ćwiczenie 5: `.*` Regex Replace – refaktor funkcji**

### 🎯 Cel:

Zamienić sygnaturę funkcji

```go
func newX() (typeA, typeB, typeC, typeD, typeE)
```

na

```go
func newX() XResult
```

### 🪄 Kroki:

1. `Ctrl + Shift + H` → Find & Replace (w pliku lub całym projekcie)
2. Włącz `.*` (regex)
3. Wpisz:

   ```
   func newX\([^)]*\)\s*\([^)]*\)
   ```
4. Replace:

   ```
   func newX() XResult
   ```
5. Sprawdź — sygnatura została podmieniona wszędzie.

---

# 💥 **Bonusowe kombinacje**

| Cel                                            | Najlepsza metoda                 |
| ---------------------------------------------- | -------------------------------- |
| Zmiana nazwy jednej zmiennej w całym pliku     | `Ctrl + F2` (rename occurrences) |
| Dodanie prefixu/sufiksu w wielu liniach        | `Alt + Shift + Drag` (kolumnowo) |
| Modyfikacja wielu fragmentów tego samego słowa | `Ctrl + D`                       |
| Refaktor wielu plików                          | `Ctrl + Shift + H` z regex       |
| Pisanie wielu linijek jednocześnie             | `Alt + Shift + ↓`                |

---

Chcesz, żebym Ci przygotował gotowy **plik ćwiczeniowy `.go` z komentarzami “TODO”**, w którym po kolei wykonujesz te ćwiczenia w VS Code (coś w stylu mini-kursu)?
