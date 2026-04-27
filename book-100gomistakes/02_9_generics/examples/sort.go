package examples

type XeFn struct {
    S       []int
    Compare func(int, int) bool
}

type SliceFn[T any] struct {
    S       []T
    Compare func(T, T) bool
}
 
func (s SliceFn[T]) Len() int           { return len(s.S) }
func (s SliceFn[T]) Less(i, j int) bool { return s.Compare(s.S[i], s.S[j]) }
func (s SliceFn[T]) Swap(i, j int)      { s.S[i], s.S[j] = s.S[j], s.S[i] }