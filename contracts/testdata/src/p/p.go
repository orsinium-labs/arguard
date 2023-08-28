package p

import "errors"

func F1(in int) error {
	if in == 0 { // want "contract: should be false: in == 0"
		return errors.New("must not be zero")
	}
	if in == 1 { // want "contract: must not be one"
		panic("must not be one")
	}
	if in == 2 { // want "contract: 42"
		panic(42)
	}
	return nil
}
