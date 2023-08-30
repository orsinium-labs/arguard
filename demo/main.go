package main

import "errors"

func target(in int) error {
	if in == 0 {
		return errors.New("must not be zero")
	}
	return nil
}

func div(n, d float64) float64 {
	if d == 0. {
		panic("denominator must not be zero")
	}
	return n / d
}

func main() {
	_ = target(0)
	x := 2.
	div(x, 0)
}
