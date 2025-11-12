package examples

type Node[T any] struct {
	Val  T
	next *Node[T]
}

func (n *Node[T]) Add(next *Node[T]) {
	n.next = next
}
