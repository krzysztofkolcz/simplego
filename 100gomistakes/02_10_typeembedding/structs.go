package typeembedding

type Foo struct {
    Bar
}
 
type Bar struct {
    Baz int
}

func Str() Foo {
	foo := Foo{}
	foo.Bar.Baz = 1
	foo.Baz = 2
	return foo
}