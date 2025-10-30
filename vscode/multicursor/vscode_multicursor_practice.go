package multicursor

import "fmt"

type typeH string
type typeB string
type typeC string
type typeD string
type typeE string

func (a typeH) doSomethingElse() {}
func (a typeB) doSomethingElse() {}
func (a typeC) doSomethingElse() {}
func (a typeD) doSomethingElse() {}
func (a typeE) doSomethingElse() {}

func main() {
	h, n, o, p, q := newH()
	h.doSomethingElse()
	n.doSomethingElse()
	o.doSomethingElse()
	p.doSomethingElse()
	q.doSomethingElse()

	fmt.Println(h, n, o, p, q)
}

func newH() (typeH, typeB, typeC, typeD, typeE) {
	var h typeH = "hello h"
	var y typeB = "hello B"
	var z typeC = "hello C"
	var k typeD = "hello D"
		var l typeE = "hello E"
	return h, y, z, k, l
}
