package p

func F1(in int) {
	if in == 0 {
		panic("must not be zero")
	}
}

func F2(in int) {
	F1(in)
	F1(1)
	F1(0) // want "contract violated: must not be zero"
}

func F3(x, y int) {
	if x == 1 {
		panic("x is one")
	}
	if y == 2 {
		panic("y is two")
	}
}

func F4(in int) {
	F3(3, 4)
	F3(3, in)
	F3(in, 3)
	F3(in, 3)
	F3(in, in)
	F3(1, in) // want "contract violated: x is one"
	F3(in, 2) // want "contract violated: y is two"
	F3(1, 2)  // want "contract violated: x is one"
}
