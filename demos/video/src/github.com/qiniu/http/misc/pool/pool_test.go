package pool

import (
	"testing"
)

// ----------------------------------------------------------

func getAll(p *Pool) (vals []int) {

	ipage := 0
	for {
		err := p.ForPage(ipage, func(v interface{}) {
			vals = append(vals, v.(int))
		})
		if err != nil {
			break
		}
		ipage++
	}
	return
}

func validAll(p *Pool, expected ...int) bool {

	vals := getAll(p)
	if len(vals) != len(expected) {
		return false
	}
	for i, v := range vals {
		if expected[i] != v {
			return false
		}
	}
	return true
}

func Test(t *testing.T) {

	p := new(Pool)

	a1 := p.Add(1)
	a2 := p.Add(3)
	a3 := p.Add(5)

	if !validAll(p, 1, 3, 5) {
		t.Fatal("p.ForEach:", getAll(p))
	}

	p.Free(a2)

	if !validAll(p, 1, 5) {
		t.Fatal("p.ForEach:", getAll(p))
	}

	p.Add(7)
	p.Free(a1)

	if !validAll(p, 7, 5) {
		t.Fatal("p.ForEach:", getAll(p))
	}

	p.Free(a3)

	if !validAll(p, 7) {
		t.Fatal("p.ForEach:", getAll(p))
	}

	p.Add(9)
	p.Add(11)

	if !validAll(p, 11, 7, 9) {
		t.Fatal("p.ForEach:", getAll(p))
	}

	for i := 0; i < N; i++ {
		p.Add(i)
	}
}

// ----------------------------------------------------------
