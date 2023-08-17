package main

import "errors"

func target(in int) error {
	if in == 0 {
		return errors.New("must not be zero")
	}
	return nil
}

func main() {
	_ = target(0)
}
