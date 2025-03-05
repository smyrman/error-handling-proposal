package xerrors_test

import "errors"

type transportModel struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type BusinessModel struct {
	A A
	B B
}

type A struct {
	a string
}

func (a A) String() string {
	return a.a
}

func ParseA(s string) (A, error) {
	l := len(s)
	if l < 3 || l > 10 {
		return A{}, errors.New("length must be in range 3-10")
	}
	return A{a: s}, nil
}

type B struct {
	b int
}

func (b B) Int() int {
	return b.b
}

func ParseB(i int) (B, error) {
	if i < 5 || i >= 100 {
		return B{}, errors.New("must be in range 5-100")
	}
	return B{b: i}, nil
}
