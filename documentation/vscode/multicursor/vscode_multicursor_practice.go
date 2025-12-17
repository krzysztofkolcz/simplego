package multicursor

import (
	"fmt"
	"log"
)

type typeA string
type typeB string
type typeC string
type typeD string
type typeE string

func (a typeA) doSomethingElse() {}
func (a typeB) doSomethingElse() {}
func (a typeC) doSomethingElse() {}
func (a typeD) doSomethingElse() {}
func (a typeE) doSomethingElse() {}

func main() {
	resultH, n, o, p, q := newH()
	resultH.doSomethingElse()
	n.doSomethingElse()
	log.Println()
	o.doSomethingElse()
	log.Println()
	p.doSomethingElse()
	log.Println()
	q.doSomethingElse()
	log.Println()

	fmt.Println(resultH, n, o, p, q)
}

func newH() (typeA, typeB, typeC, typeD, typeE) {
	var resultH typeA = "helllo resultH"
	var y typeB = "helllo B"
	var z typeC = "helllo C"
	var k typeD = "helllo D"
	var l typeE = "helllo E"
	return resultH, y, z, k, l
}
