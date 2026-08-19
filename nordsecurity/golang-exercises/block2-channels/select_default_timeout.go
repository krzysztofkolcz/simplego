package block2

// Zadanie 26 — `select` z `default` (non-blocking) i z `time.After` (timeout).
//
// Teoria:
//
// `select` wygląda jak `switch`, ale każdy `case` to operacja na kanale
// (send albo receive). Działanie:
//   - Jeśli DOKŁADNIE JEDEN case jest gotowy (kanał ma coś do odebrania /
//     jest miejsce żeby wysłać) — `select` go wykonuje.
//   - Jeśli WIĘCEJ NIŻ JEDEN case jest gotowy naraz — Go wybiera LOSOWO
//     jeden z nich (celowo, żeby nikt nie polegał na kolejności `case`ów).
//   - Jeśli ŻADEN case nie jest gotowy — `select` BLOKUJE się i czeka, aż
//     którykolwiek się uaktywni. To jest domyślne zachowanie (tak jak zwykłe
//     `<-ch` blokuje).
//
// `default`:
//
//	select {
//	case v := <-ch:
//	    // odebrano
//	default:
//	    // ch nie miał nic gotowego DOKŁADNIE TERAZ — zamiast czekać,
//	    // od razu leci tutaj
//	}
//
// `default` sprawia, że `select` przestaje blokować — sprawdza kanały RAZ,
// i jeśli żaden nie jest gotowy, natychmiast wykonuje `default`. To jest
// standardowy sposób na "spróbuj odebrać/wysłać, ale się nie zatrzymuj".
//
// `time.After(d)`:
//
//	select {
//	case v := <-ch:
//	    // odebrano na czas
//	case <-time.After(d):
//	    // minęło `d` i nikt nic nie wysłał — timeout
//	}
//
// `time.After(d)` zwraca kanał, do którego coś "wpadnie" dopiero po czasie
// `d`. Użyty jako case w `select`, działa jak timeout na operację na kanale:
// `select` i tak normalnie by zablokował się w nieskończoność czekając na
// `ch`, ale teraz ma alternatywę, która "obudzi" go po `d`, jeśli `ch` się
// nie odezwie wcześniej.
//
// Różnica `default` vs `time.After`: `default` to "sprawdź RAZ, teraz,
// zero czekania". `time.After` to "poczekaj, ale nie dłużej niż `d`".

import "time"

// TryReceiveNonBlocking próbuje odebrać jedną wartość z ch BEZ blokowania.
// Jeśli w kanale jest coś gotowe do odebrania — zwraca (wartość, true).
// Jeśli nic nie jest gotowe — zwraca natychmiast (0, false), nie czekając
// ani chwili.
func TryReceiveNonBlocking(ch chan int) (value int, ok bool) {
	select {
	case value := <-ch:
		return value, true
	default:
		return 0, false
	}
}

// ReceiveWithTimeout czeka na wartość z ch, ale nie dłużej niż d. Jeśli
// coś przyjdzie z ch w tym czasie — zwraca (wartość, false). Jeśli minie d
// zanim cokolwiek przyjdzie — zwraca (0, true) (timedOut).
func ReceiveWithTimeout(ch chan int, d time.Duration) (value int, timedOut bool) {
	select {
	case value := <-ch:
		return value, false
	case <-time.After(d):
		return 0, true
	}
}
