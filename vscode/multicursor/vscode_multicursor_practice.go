package multicursor

import "fmt"

type typeAA string
type typeBB string
type typeCC string
type typeDD string
type typeEE string

func (a typeAA) doSomething() {}
func (a typeBB) doSomething() {}
func (a typeCC) doSomething() {}
func (a typeDD) doSomething() {}
func (a typeEE) doSomething() {}

func main() {
	haa, na, oa, pa, qa := newH()
	haa.doSomething()
	na.doSomething()
	oa.doSomething()
	pa.doSomething()
	qa.doSomething()

	fmt.Println(haa, na, oa, pa, qa)
}

func newH() (typeAA, typeBB, typeCC, typeDD, typeEE) {
	var h typeAA = "hello h"
	var y typeBB = "hello B"
	var z typeCC = "hello C"
	var k typeDD = "hello D"
	var l typeEE = "hello E"
	return h, y, z, k, l
}
