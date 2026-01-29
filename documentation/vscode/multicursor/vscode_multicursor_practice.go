package multicursor

import (
	"fmt"
	// "log"
)

type typeA string
type typeB string
type typeC string
type typeD string
type typeE string

type typeX struct{
	a typeA
	b typeB
	c typeC
	d typeD
	e typeE

}

func (a typeA) doSomething() {}
func (a typeB) doSomething() {}
func (a typeC) doSomething() {}
func (a typeD) doSomething() {}
func (a typeE) doSomething() {}

func main() {
	ee, aa, bb, cc, dd := newH()
	ee.doSomething()
	aa.doSomething()
	fmt.Println()
	bb.doSomething()
	fmt.Println()
	cc.doSomething()
	fmt.Println()
	dd.doSomething()
	fmt.Println()

	fmt.Println(ee, aa, bb, cc, dd)
	resultX := typeX{
		a: ee,
		b: aa,
		c: bb,
		d: cc,
		e: dd,
	}
	fmt.Println(resultX)
}

func newH() (typeA, typeB, typeC, typeD, typeE) {
	var p typeA = "helllo p"
	var y typeB = "helllo B"
	var z typeC = "helllo C"
	var k typeD = "helllo D"
	var l typeE = "helllo E"
	return p, y, z, k, l
}
