package examples

type Foo struct{}

// Not compile
// func (Foo) bar[T any](t T) {}
// ...: methods cannot have type parameters
