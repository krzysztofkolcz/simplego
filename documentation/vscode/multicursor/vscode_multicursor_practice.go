package multicursor

import "fmt"

type typeA string
type typeB string
type typeC string
type typeD string
type typeE string

func (a typeA) doSomething() {}
func (a typeB) doSomething() {}
func (a typeC) doSomething() {}
func (a typeD) doSomething() {}
func (a typeE) doSomething() {}

func main() {
	h, n, o, p, q := newH()
	h.doSomething()
	n.doSomething()
	o.doSomething()
	p.doSomething()
	q.doSomething()

	fmt.Println(h, n, o, p, q)
}

func newH() (typeA, typeB, typeC, typeD, typeE) {
	var h typeA = "helllo h"
	var y typeB = "helllo B"
	var z typeC = "helllo C"
	var k typeD = "helllo D"
	var l typeE = "helllo E"
	return h, y, z, k, l
}
