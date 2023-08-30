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

func F2(in int) error {
	if in == 4 {
		println(in)
	}
	return nil
}

func F3(in int) error {
	if 12 == 12 { // nolint: staticcheck
		panic("always reachable")
	}
	return nil
}

func F4(in int) error {
	if true { // nolint: staticcheck
		panic("always reachable")
	}
	return nil
}

func F5(in int) error { // nolint: staticcheck
	in = 12 // nolint: staticcheck
	if in == 12 {
		panic("always reachable")
	}
	return nil
}
