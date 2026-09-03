package block2

import "context"

// Zadanie 28 — Pipeline z anulowaniem przez context.Context.
//
// Teoria:
//
// Pipeline to kilka etapów połączonych kanałami: każdy etap czyta z kanału
// wejściowego, coś robi, wysyła wynik do kanału wyjściowego, i (tak jak
// w fan-out/fan-in z zadania 27) zamyka swój output, gdy jego input się
// zamknie i wyczerpie.
//
// Problem: co jeśli konsument na końcu pipeline'u przestaje odbierać,
// zanim wcześniejsze etapy skończą produkować? Bez żadnego mechanizmu,
// wcześniejsze etapy blokują się na wysyłaniu DO KOŃCA PROGRAMU — to jest
// goroutine leak (potwierdziliśmy to przed chwilą).
//
// Rozwiązanie: KAŻDY etap, zamiast zwykłego `out <- v`, robi:
//
//	select {
//	case out <- v:
//	    // wysłano normalnie
//	case <-ctx.Done():
//	    // konsument (albo cokolwiek innego) anulował ctx — przestań
//	    // produkować i zakończ się (zamknij output, wróć z funkcji)
//	}
//
// To samo dotyczy odbierania, jeśli etap sam czeka na coś z wejścia poza
// zwykłym `range` (tutaj używamy `range`, więc to nie jest potrzebne przy
// odbiorze — `range` i tak kończy się gdy input zostanie zamknięty).
//
// Dlaczego context.Context, a nie osobny kanał `done chan struct{}`?
// - ctx propaguje się drzewiasto przez wiele funkcji/warstw bez przekazywania
//   osobnego kanału do każdej z osobna,
// - `context.WithCancel`/`WithTimeout`/`WithDeadline` dają gotowe sposoby
//   automatycznego anulowania (nie tylko ręcznego), a `ctx.Done()` zwraca
//   ten sam typ kanału niezależnie od tego, co go anulowało,
// - to idiomatyczny, powszechnie oczekiwany sposób w Go (biblioteka
//   standardowa, gRPC, http.Request — wszędzie ctx).
//
// Zadanie:
//
// Zbuduj 3-etapowy pipeline z dwóch funkcji poniżej (generator jest już
// gotowy jako przykład wzorca):
//
//  1. Generator (gotowy) — wysyła kolejne liczby z `nums` do kanału,
//     zamyka go na końcu, respektuje ctx.Done().
//  2. Square — dopisz: czyta z `in` (przez `range`), podnosi każdą liczbę
//     do kwadratu, wysyła do `out` przez select+ctx.Done(), zamyka `out`
//     na końcu.
//
// Generator pokazuje wzorzec — Square ma wyglądać bardzo podobnie.

// Generator wysyła kolejne wartości z nums do zwróconego kanału. Zamyka
// kanał, gdy skończy wysyłać WSZYSTKIE wartości, albo wcześniej, jeśli
// ctx zostanie anulowany w trakcie wysyłania.
func Generator(ctx context.Context, nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			select {
			case out <- n:
				// wysłano, idziemy do kolejnej liczby
			case <-ctx.Done():
				// anulowano — przestań produkować
				return
			}
		}
	}()
	return out
}

// Square czyta liczby z in, podnosi każdą do kwadratu i wysyła do
// zwróconego kanału. Respektuje ctx.Done() przy wysyłaniu (tak jak
// Generator). Zamyka swój output, gdy in się zamknie i wyczerpie, albo
// wcześniej przy anulowaniu ctx.
func Square(ctx context.Context, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer func() {
			close(out)
		}()
		for i := range in {
			isqr := i * i
			select {
			case out <- isqr:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
