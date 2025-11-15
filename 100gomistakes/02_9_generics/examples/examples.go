package examples

import "strconv"

func foo[T any](t T) {

}

type customConstraint interface {
	~int | ~string
}

type CustomConstraint2 interface {
	~int
	String() string
}

// implements customConstraint2
type CustomInt int

func (i CustomInt) String() string {
	return strconv.Itoa(int(i))
}

func GetKeys[K CustomConstraint2,
	V any](m map[K]V) []K {
	var keys []K
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestGetKeys2() []CustomInt {
	m := map[CustomInt]int{
		1: 1,
		2: 2,
		3: 3,
	}
	keys := GetKeys(m)
	return keys
}
