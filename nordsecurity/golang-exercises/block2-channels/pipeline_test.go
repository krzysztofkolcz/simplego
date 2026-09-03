package block2

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	ctx := context.Background()
	items := []int{}
	result := []int{}
	expected := []int{}
	for i := 0; i < 100; i++ {
		items = append(items, i)
	}

	for i := 0; i < 100; i++ {
		expected = append(expected, i*i)
	}
	t.Run("pipeline test", func(t *testing.T) {

		genout := Generator(ctx, items...)
		sqout := Square(ctx, genout)
		for s := range sqout {
			result = append(result, s)
		}

		assert.ElementsMatch(t, expected, result)
	})
}

/*
Napisz drugi test, TestPipelineCancellation, sprawdzający dokładnie ten scenariusz:

1. ctx, cancel := context.WithCancel(context.Background()).
2. Puść pipeline z dużą liczbą elementów (np. 1000) — żeby generator na pewno nie zdążył się wyczerpać sam.
3. Odbierz z sqout tylko 1-2 wartości (<-sqout), potem wywołaj cancel().
4. Sprawdź, że sqout faktycznie się zamyka po cancel() (czyli pipeline się kończy), zamiast wisieć w nieskończoność.

Podpowiedź do punktu 4 — jak sprawdzić "zamyka się, a nie wisi" w teście, żeby test się nie zawiesił na zawsze, jeśli coś jest nie tak: użyj select z case <-time.After(...) (dokładnie wzorzec z zadania 26) opakowującego kolejny odbiór z sqout, i t.Fatal(...) w gałęzi timeout.

Spróbuj napisać ten test, to jest właściwy dowód, że ctx.Done() w Twoim Square/Generator naprawdę robi to, co ma robić.
*/
func TestPipelineCancelation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	items := []int{}
	for i := 0; i < 10000; i++ {
		items = append(items, i)
	}

	t.Run("pipeline cancelation", func(t *testing.T) {
		genout := Generator(ctx, items...)
		sqout := Square(ctx, genout)
		<-sqout
		<-sqout
		cancel()

		// Po cancel() sqout powinien się zamknąć w rozsądnym czasie.
		// Może po drodze przyjść jeszcze jedna wartość "w locie" (select
		// w Square losowo wybrał `out <- isqr` zamiast `ctx.Done()`) —
		// to normalne, więc czytamy w pętli, aż zobaczymy ok == false.
		deadline := time.After(1 * time.Second)
		for {
			select {
			case _, ok := <-sqout:
				if !ok {
					return // kanał zamknięty — sukces, pipeline się zakończył
				}
				// dostaliśmy jeszcze jedną wartość "w locie" — czytamy dalej
			case <-deadline:
				t.Fatal("sqout nie zamknął się po cancel() — możliwy goroutine leak")
			}
		}
	})
}
