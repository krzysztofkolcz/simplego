package block2

import "sync"

// Zadanie 27 — Fan-out / fan-in.
//
// Teoria:
//
// "Fan-out": N workerów (goroutine) czyta z JEDNEGO wspólnego kanału
// wejściowego. Każdy job trafia tylko do JEDNEGO workera (kanały nie
// duplikują danych) — to naturalny load-balancing, bo szybszy worker
// po prostu częściej zdąży odebrać kolejny job.
//
// "Fan-in": ci sami workerzy piszą wyniki do JEDNEGO wspólnego kanału
// wyjściowego. Kilka niezależnych źródeł "zbiega się" w jeden strumień.
//
// Kluczowy problem do rozwiązania: KTO i KIEDY zamyka kanał wyjściowy?
//   - Konsument (ten kto robi `range output`) będzie czekał w nieskończoność,
//     jeśli nikt nie zamknie `output` — `range` po kanale kończy się TYLKO
//     gdy kanał jest zamknięty (i pusty).
//   - Nie może zamknąć go pierwszy worker, który skończy — inni workerzy
//     wciąż próbowaliby wysyłać do zamkniętego kanału → panic
//     "send on closed channel".
//   - Trzeba więc: policzyć, kiedy WSZYSCY workerzy skończyli (do tego
//     służy `sync.WaitGroup`), i DOPIERO WTEDY zamknąć `output`.
//   - Ale `wg.Wait()` blokuje — nie możemy go wywołać w tej samej
//     goroutine, która ma zaraz zwrócić `output` konsumentowi (bo
//     nigdy by do tego nie doszło). Rozwiązanie: osobna, dodatkowa
//     goroutine, której JEDYNYM zadaniem jest: `wg.Wait()`, a potem
//     `close(output)`.
//
// Zadanie:
//
// Zaimplementuj FanOutFanIn poniżej:
//   - Odpal `numWorkers` goroutine. Każda w pętli `for job := range input`
//     woła `worker(job)` i wysyła wynik do `output`.
//   - Każdy worker, kończąc (bo `input` zostało zamknięte i osuszone),
//     musi to zgłosić przez WaitGroup.
//   - Dodatkowa goroutine czeka aż WSZYSCY workerzy skończą, i dopiero
//     wtedy zamyka `output`.
//   - Funkcja NIE czeka na nic sama — zwraca `output` od razu, konsument
//     sam zrobi `range output` (lub `<-output` w pętli) i to zablokuje się
//     naturalnie, aż faktycznie coś przyjdzie albo kanał się zamknie.
func FanOutFanIn(input <-chan int, numWorkers int, worker func(int) int) <-chan int {
	output := make(chan int)
	in := make(chan int)
	// TODO: zaimplementuj
	var wg sync.WaitGroup
	_ = wg
	for range MaxWorkers {
		go fworker(in, output)
	}
	return output
}

const MaxWorkers = 10

func fworker(in <-chan int, out chan<- int) {
	for r := range in {
		out <- process(r)
	}
}

func process(r int) int {
	return r
}
