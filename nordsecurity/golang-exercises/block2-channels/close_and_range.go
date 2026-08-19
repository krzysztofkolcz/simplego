package block2

import "fmt"

/*
close() nie kasuje danych, które są już w buforze. Kanał to jak kolejka: close() tylko oznacza "nikt już nic więcej nie wyśle", ale to co już jest w buforze — zostaje i wciąż można to odebrać. Dopiero gdy bufor jest pusty i kanał jest zamknięty, kolejny odczyt zwraca zero value + ok == false.

Więc dla naszego przykładu:
1. range odbierze 1, 2, 3 — normalnie, jak z każdego kanału.
2. Po odebraniu trzeciej wartości, range próbuje odebrać czwartą — bufor jest pusty, kanał jest zamknięty, więc range automatycznie kończy pętlę (to jest właśnie mechanizm range po kanale: sam sprawdza ok za kulisami i przerywa, gdy ok == false).

Czyli output będzie:
odebrano: 1
odebrano: 2
odebrano: 3
pętla range się skończyła

Zostaje jeszcze Pytanie 2 od Ciebie — a właścm powyżej (pętla kończy się sama, bez panic).To odróżnia range/v, ok := <-ch (bezpieczne) anału (ch <- x po close(ch)), co jest zupełnie inną historią. Chcesz zgadnąć, co się stanie  to napiszemy?
*/
func ProduceAndRange() []int {
	ch := make(chan int, 10)
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch)
	items := make([]int, 0, 3)
	for value := range ch {
		items = append(items, value)
	}
	return items
}

/*
wysyłanie do zamkniętego kanału:
panic: send on closed channel.
To jest twarda reguła: tylko producent (ten kto wysyła) może zamykać kanał, i nigdy nie wysyła nic po close().
Konsument nigdy nie powinien zamykać kanału, którego nie jest właścicielem.
*/
func SendToClosedChannel() (panicked bool, msg string) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			msg = fmt.Sprint(r)
		}
	}()
	ch := make(chan int, 10)
	close(ch)
	ch <- 1
	return false, ""
}
