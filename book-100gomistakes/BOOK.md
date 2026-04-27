https://livebook.manning.com/book/100-go-mistakes-and-how-to-avoid-them/chapter-2#315

# 2.9 #9
Konstrukcja generyka:
```
func foo[T any](t T) {
    // ...
}
```

Ograniczenie typu do int lub string:

```
type customConstraint interface {
    ~int | ~string
}
 
func getKeys[K customConstraint, V any](m map[K]V) []K {
    // Same implementation
}
```

## ~int
ograniczenie do typów, których zasadniczym typem jest int
np. określenie typu w ten sposób:

```
type customConstraint interface {
    ~int
    String() string
}
```

```
type customInt int
 
func (i customInt) String() string {
    return strconv.Itoa(int(i))
}
```
customInt spełnia customConstraint i zasadniczym typem customInt jest int.

jeżeli określone byłoby bez tyldy:
```
type customConstraint interface {
    int
    String() string
}
```
customInt nie spełniałby tego interfejsu.