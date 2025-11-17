package typeembedding

import "sync"

type WrongInMem struct {
	sync.Mutex // Wrong! it is possible to access Lock
	m          map[string]int
}

type OKInMem struct {
	mu sync.Mutex
	m  map[string]int
}

func New() *WrongInMem {
	return &WrongInMem{m: make(map[string]int)}
}

func (i *WrongInMem) Get(key string) (int, bool) {
	i.Lock()
	v, contains := i.m[key]
	i.Unlock()
	return v, contains
}

func Wrong() {
	m := WrongInMem{}
	m.Lock() // ??
}

func (r *OKInMem) Blabla(a string, b int) error {
	return nil
}
